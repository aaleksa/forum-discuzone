package handlers

import (
	"database/sql"
	"fmt"
	"forum/internal"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"io"

	_ "github.com/mutecomm/go-sqlcipher/v4"
	"net/http"
	"os"
)

func UploadAvatarHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil || user == nil {
			errors.RenderError(w, http.StatusUnauthorized, "Unauthorized", "Login required.")
			return
		}

		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}

		file, handler, err := r.FormFile("avatar")
		if err != nil {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "No file uploaded.")
			return
		}
		defer file.Close()

		// Save file
		filename := fmt.Sprintf("%d_%s", user.ID, handler.Filename)
		filePath := "static/uploads/" + filename
		out, err := os.Create(filePath)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Error", "Failed to save avatar.")
			return
		}
		defer out.Close()
		io.Copy(out, file)

		_, err = db.Exec("UPDATE users SET avatar_url = ? WHERE id = ?", "/static/uploads/"+filename, user.ID)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Error", "Failed to update avatar URL.")
			return
		}

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}
