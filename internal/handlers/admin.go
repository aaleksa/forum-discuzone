package handlers

import (
	"database/sql"
	"forum/internal/models"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"net/http"
	"text/template"
)

func AdminUsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, _ := utils.GetUserFromSession(w, r, db)
		if currentUser == nil || currentUser.Role != "admin" {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		users, err := utils.GetAllUsers(db)
		if err != nil {
			log.Printf("Error getting users: %v", err)
			http.Error(w, "Error loading users", http.StatusInternalServerError)
			return
		}

		modRequests, err := utils.GetModerationRequests(db)
		if err != nil {
			log.Printf("Error getting moderation requests: %v", err)
			http.Error(w, "Error loading moderation requests", http.StatusInternalServerError)
			return
		}

		log.Printf("Loaded %d moderation requests", len(modRequests))

		data := struct {
			CurrentUser        *models.User
			Users              []models.User
			ModerationRequests []models.ModerationRequest
		}{
			CurrentUser:        currentUser,
			Users:              users,
			ModerationRequests: modRequests,
		}

		tmpl := template.Must(template.ParseFiles(
			"templates/layout_admin.html",
			"templates/header_auth.html",
			"templates/admin_users.html",
			"templates/nav_admin.html",
		))

		err = tmpl.ExecuteTemplate(w, "layout", data)
		if err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}

func BanUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil || !utils.IsAdmin(user) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID := r.FormValue("user_id")
		if userID == "" {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE users SET banned = 1 WHERE id = ?", userID)
		if err != nil {
			http.Error(w, "Failed to ban user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

func UnbanUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil || !utils.IsAdmin(user) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID := r.FormValue("user_id")
		if userID == "" {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE users SET banned = FALSE WHERE id = ?", userID)
		if err != nil {
			http.Error(w, "Failed to unban user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}
