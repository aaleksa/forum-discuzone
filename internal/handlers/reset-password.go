package handlers

import (
	"database/sql"
	"forum/internal"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"text/template"
)

func ResetPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")

		var exists bool
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM password_resets WHERE token = ? AND expires_at > CURRENT_TIMESTAMP)`, token).Scan(&exists)
		if err != nil || !exists {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid or expired token.")
			return
		}

		data := struct {
			Token string
		}{
			Token: token,
		}

		tmpl, err := template.ParseFiles(
			"templates/layout_auth.html",
			"templates/reset_password.html",
			"templates/header_auth.html",
		)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template not loaded.")
			return
		}

		err = tmpl.ExecuteTemplate(w, "layout", data)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template execution error.")
		}
		tmpl.Execute(w, struct{ Token string }{Token: token})
	}
}

func ResetPasswordSubmitHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		newPassword := r.FormValue("password")

		var userID int
		err := db.QueryRow(`SELECT user_id FROM password_resets WHERE token = ? AND expires_at > CURRENT_TIMESTAMP`, token).Scan(&userID)
		if err != nil {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid or expired token.")
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Could not hash password.")
			return
		}

		_, err = db.Exec(`UPDATE users SET password = ? WHERE id = ?`, string(hashedPassword), userID)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Could not update password.")
			return
		}

		_, _ = db.Exec(`DELETE FROM password_resets WHERE token = ?`, token)
		// ✅ Return 200 OK for fetch() JS
		w.WriteHeader(http.StatusOK)
		defer db.Close()
	}
}
