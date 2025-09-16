package handlers

import (
	"database/sql"

	"forum/internal"
	"forum/internal/models"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"net/http"
	"text/template"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func HandlerProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the current user
		user, err := utils.GetUserFromSession(w, r, db)
		log.Printf("user: %+v", user)
		if err != nil {
			log.Printf("Error getting user from session: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to get user data")
			return
		}
		if user == nil {
			errors.RenderError(w, http.StatusUnauthorized, "Unauthorized", "Please log in to view this page")
			return
		}
		// Get current user from session
		currentUser, _ := utils.GetUserFromSession(w, r, db) // Ignore error if not logged in

		// We get a list of posts from the database
		posts, err := GetPosts(db, currentUser)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Error retrieving posts.")
			return
		}

		// Receiving notifications
		notifications, err := GetAllNotifications(db, user.ID)
		if err != nil {
			log.Printf("Error receiving notifications: %v", err)
		}

		data := models.ProfilePageData{
			User:             *user,
			CurrentUser:      user,
			Notifications:    notifications,
			PostsWithComment: posts,
		}

		tmpl, err := template.ParseFiles(
			"templates/layout.html",
			"templates/header.html",
			"templates/nav.html",
			"templates/profile.html",
			"templates/notifications.html",
			"templates/user_comments.html",
			"templates/notification_list.html",
		)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template loading error")
			return
		}

		err = tmpl.ExecuteTemplate(w, "layout", data)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Render error: "+err.Error())
		}
	}
}
