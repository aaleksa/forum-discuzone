package handlers

import (
	"database/sql"
	"encoding/json"
	"forum/internal"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"net/http"
)

func PromoteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errors.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed", "Method not allowed")
			return
		}

		currentUser, _ := utils.GetUserFromSession(w, r, db)
		if currentUser == nil || currentUser.Role != "admin" {
			errors.RenderError(w, http.StatusForbidden, "Forbidden", "Forbidden")
			return
		}

		userID := r.FormValue("userID")
		newRole := r.FormValue("role")
		_, err := db.Exec("UPDATE users SET role = ? WHERE id = ?", newRole, userID)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to update role.")

			return
		}

		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

func ApproveModeratorRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			errors.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed", "Method not allowed")
			return
		}

		currentUser, _ := utils.GetUserFromSession(w, r, db)
		if currentUser == nil || currentUser.Role != "admin" {
			errors.RenderError(w, http.StatusForbidden, "Forbidden", "Forbidden")
			return
		}

		requestID := r.FormValue("requestID")
		if requestID == "" {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Missing request ID")
			return
		}

		// Update the status of the request to "approved"
		_, err := db.Exec(`
        UPDATE moderator_requests 
        SET status = 'approved', 
            reviewed_at = CURRENT_TIMESTAMP, 
            reviewed_by = ?
        WHERE id = ? AND status = 'pending'
    `, currentUser.ID, requestID)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to approve request")
			return
		}

		// We get the user_id from the request
		var userID int
		err = db.QueryRow("SELECT user_id FROM moderator_requests WHERE id = ?", requestID).Scan(&userID)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to get user ID")
			return
		}

		// Update the user role to "moderator"
		_, err = db.Exec("UPDATE users SET role = 'moderator' WHERE id = ?", userID)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to update user role")
			return
		}

		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

func RejectModeratorRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errors.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed", "Method not allowed")
			return
		}

		currentUser, _ := utils.GetUserFromSession(w, r, db)
		if currentUser == nil || currentUser.Role != "admin" {
			errors.RenderError(w, http.StatusForbidden, "Forbidden", "Forbidden")
			return
		}

		requestID := r.FormValue("requestID")
		if requestID == "" {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Missing request ID")
			return
		}
		// Update the status of the request to "rejected"
		_, err := db.Exec(`
        UPDATE moderator_requests 
        SET status = 'rejected', 
            reviewed_at = CURRENT_TIMESTAMP, 
            reviewed_by = ?
        WHERE id = ? AND status = 'pending'
    `, currentUser.ID, requestID)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to reject request")
			return
		}

		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

func CheckModeratorStatusHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errors.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed", "Method not allowed")
			return
		}

		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil || user == nil {
			errors.RenderError(w, http.StatusUnauthorized, "Unauthorized", "Please log in")
			return
		}

		var role string

		// ✅ Check user role
		err = db.QueryRow(`
        SELECT role FROM users WHERE id = ?
    `, user.ID).Scan(&role)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Database error", "Failed to check user role")
			return
		}

		if role == "moderator" {
			// 🔁 User is already a moderator - revert immediately
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status": "approved",
			})
			return
		}

		// 🔄 If not a moderator, check the status of the last request
		var status string
		err = db.QueryRow(`
        SELECT status FROM moderator_requests
        WHERE user_id = ?
        ORDER BY requested_at DESC
        LIMIT 1
    `, user.ID).Scan(&status)

		if err != nil {
			if err == sql.ErrNoRows {
				// Didn't send a request
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"status": "not_requested",
				})
				return
			}
			errors.RenderError(w, http.StatusInternalServerError, "Database error", "Failed to check status")
			return
		}

		// Return the found request status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": status,
		})
	}
}
