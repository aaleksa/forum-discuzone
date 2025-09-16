package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/internal/models"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"net/http"
	"text/template"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func HandleRequestModerator(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil || user.Role != "user" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		err = SendModeratorRequest(db, user.ID)

		// --- 👇 Check: is this an AJAX request?
		isAjax := r.Header.Get("X-Requested-With") == "XMLHttpRequest"

		if isAjax {
			w.Header().Set("Content-Type", "application/json")
			if err != nil {
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": err.Error(),
				})
			} else {
				// status can be "pending", "approved", "rejected" - it depends on your logic
				json.NewEncoder(w).Encode(map[string]string{
					"status": "pending",
				})
			}
			return
		}

		profileData := models.ProfilePageData{CurrentUser: user}
		if err != nil {
			profileData.RequestError = err.Error()
		} else {
			profileData.RequestSent = true
		}

		tmpl := template.Must(template.ParseFiles("templates/profile.html"))
		tmpl.ExecuteTemplate(w, "profile", profileData)
	}
}

func SendModeratorRequest(db *sql.DB, userID int) error {
	// Check if there is already a request from this user
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM moderator_requests WHERE user_id = ? AND status = 'pending')", userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("The request could not be verified: %v", err)
	}

	if exists {
		return fmt.Errorf("Request has already been sent and is awaiting review")
	}

	// Вставка нового запиту
	_, err = db.Exec("INSERT INTO moderator_requests (user_id, status, requested_at) VALUES (?, 'pending', datetime('now'))", userID)
	if err != nil {
		return fmt.Errorf("Failed to send request: %v", err)
	}

	return nil
}
