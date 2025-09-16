package handlers

import (
	"database/sql"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"net/http"
)

func CreateCategoryHandler(db *sql.DB) http.HandlerFunc {
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

		name := r.FormValue("category_name")
		if name == "" {
			http.Error(w, "Category name required", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("INSERT INTO categories (name) VALUES (?)", name)
		if err != nil {
			http.Error(w, "Failed to create category", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
	}
}
