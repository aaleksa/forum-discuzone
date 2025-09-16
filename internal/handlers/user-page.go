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
)

func HandlerUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("DEBUG: Proccessing in HandlerUser")

		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil {
			log.Println("DEBUG: MustGetUserID failed in GetUserFromSession")
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Session check error=====.")
			return
		}

		if user == nil {
			log.Println("DEBUG: User is nil in HandlerUser")
			errors.RenderError(w, http.StatusUnauthorized, "Unauthorized", "No user session found.")
			return
		}
		log.Println("DEBUG: User is not  nil in HandlerUser: %v", user)

		// We get a list of posts from the database
		posts, err := GetPosts(db, user)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Error retrieving posts.")
			return
		}

		// Get categories
		categories, err := utils.GetCategories(w, db)
		if err != nil {
			log.Printf("ERROR: Failed to get categories: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Something went wrong.")
			return
		}
		log.Println("DEBUG: Fetched categories-user-page")

		data := models.UserPageData{
			Posts:       posts,
			CurrentUser: user,
			Categories:  categories,
		}

		// Download the post creation page template

		tmpl, err := template.ParseFiles(
			"templates/layout.html",
			"templates/user_page.html",
			"templates/header.html",
			"templates/nav.html",
			"templates/post_list.html",
			"templates/post_list_item.html",
			"templates/filters.html",
			"templates/notifications.html",
		)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template not load.")
			return
		}
		tmpl.ExecuteTemplate(w, "layout", data)
	}
}
