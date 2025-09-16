package handlers

import (
	"database/sql"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"net/http"
)

func DeleteCategoryHandler(db *sql.DB) http.HandlerFunc {
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

		id := r.FormValue("category_id")
		if id == "" {
			http.Error(w, "Category ID required", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("DELETE FROM categories WHERE id = ?", id)
		if err != nil {
			http.Error(w, "Failed to delete category", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
	}
}
