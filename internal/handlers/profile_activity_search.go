package handlers

import (
	"database/sql"
	"forum/internal"
	"forum/internal/models"
	"forum/internal/utils"
	"html/template"
	"log"
	"net/http"
	"strings"
	// "time"
	// "context"
	"fmt"
)

func HandlerUserActivitySearch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user session
		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil {
			log.Printf("Session error: %v", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if user == nil {
			log.Println("No user in session")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get query parameters
		query := strings.TrimSpace(r.URL.Query().Get("query"))
		activityType := strings.TrimSpace(r.URL.Query().Get("type"))

		// // Create template with context timeout
		// ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		// defer cancel()

		// Parse templates
		tmpl := template.New("base").Funcs(template.FuncMap{
			"formatDate": utils.FormatDate,
		})

		tmpl, err = tmpl.ParseFiles(
			"templates/layout.html",
			"templates/header.html",
			"templates/nav.html",
			"templates/profile.html",
			"templates/notifications.html",
			"templates/user_comments.html",
			"templates/notification_list.html",
			"templates/profile_activity_search.html",
		)
		if err != nil {
			log.Printf("Template parsing error: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Template Error", "Could not load page templates.")
			return
		}

		// Get current user from session
		currentUser, _ := utils.GetUserFromSession(w, r, db) // Ignore error if not logged in

		// Handle empty query
		if query == "" {
			data := models.ActivitySearchPageData{
				Query:            "",
				Message:          "Please enter a search query.",
				User:             *user,
				CurrentUser:      user,
				Results:          models.UserActivityResults{},
				Notifications:    []models.Notification{},
				PostsWithComment: []models.PostView{},
			}

			if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
				log.Printf("Template execution error: %v", err)
				errors.RenderError(w, http.StatusInternalServerError, "Render Error", "Could not display search page.")
			}
			return
		}

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

		// Perform search with context
		results, err := utils.SearchUserActivity(db, user.ID, query, activityType)
		if err != nil {
			log.Printf("Search failed for user %d: %v", user.ID, err)
			errors.RenderError(w, http.StatusInternalServerError, "Search Error", "Could not complete search.")
			return
		}

		// Prepare response data
		data := models.ActivitySearchPageData{
			Query:            query,
			FilterType:       activityType,
			Results:          results,
			User:             *user,
			CurrentUser:      user,
			Notifications:    notifications,
			PostsWithComment: posts,
		}

		if len(results.Posts) == 0 && len(results.Comments) == 0 && len(results.Likes) == 0 {
			data.Message = fmt.Sprintf("No results found for '%s'", query)
		}

		// Execute template
		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
			log.Printf("Template execution failed: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Render Error", "Could not display results.")
		}
	}
}
