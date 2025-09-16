package handlers

import (
	"database/sql"
	"forum/internal"
	"forum/internal/models"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"net/http"
	"text/template"
)

type Category struct {
	ID   int
	Name string
}

func GetCategoriesForPost(db *sql.DB, postID int) ([]Category, error) {
	query := `
        SELECT c.id, c.name
        FROM categories c
        JOIN post_categories pc ON c.id = pc.category_id
        WHERE pc.post_id = $1`
	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func AdminCategoriesPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil || !utils.IsAdmin(user) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		rows, err := db.Query("SELECT id, name FROM categories ORDER BY name ASC")
		if err != nil {
			http.Error(w, "Failed to load categories", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var categories []Category
		for rows.Next() {
			var c Category
			if err := rows.Scan(&c.ID, &c.Name); err == nil {
				categories = append(categories, c)
			}
		}

		// Create a data structure for the template
		data := struct {
			Categories  []Category
			CurrentUser *models.User
		}{
			Categories:  categories,
			CurrentUser: user,
		}

		tmpl, err := template.ParseFiles(
			"templates/layout_admin.html",
			"templates/header.html",
			"templates/nav_admin.html",
			"templates/admin_categories.html",
		)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template not loaded.")
			return
		}

		err = tmpl.ExecuteTemplate(w, "layout", data)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template execution failed.")
		}
	}
}
