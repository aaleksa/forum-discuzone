package handlers

import (
	"database/sql"
	"forum/internal"
	"forum/internal/models"
	"forum/internal/utils"
	"html/template"
	"net/http"
	"strings"
)

func HandlerSearch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("query"))

		tmpl := template.Must(template.ParseFiles(
			"templates/layout.html",
			"templates/nav.html",
			"templates/header.html",
			"templates/search_results.html",
			"templates/post_list_item.html",
			"templates/notifications.html",
		))

		// 🧑 Get current user (if available)
		user, _ := utils.GetUserFromSession(w, r, db) // ignore error for optional login

		if query == "" {
			tmpl.ExecuteTemplate(w, "layout", models.SearchPageData{
				Query:       "",
				Results:     nil,
				Message:     "Please enter a search query.",
				CurrentUser: user,
			})
			return
		}

		posts, err := utils.SearchPosts(db, query)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Error performing search.")
			return
		}

		message := ""
		if len(posts) == 0 {
			message = "No results found for \"" + query + "\"."
		}

		tmpl.ExecuteTemplate(w, "layout", models.SearchPageData{
			Query:       query,
			Results:     posts,
			Message:     message,
			CurrentUser: user,
		})
	}
}
