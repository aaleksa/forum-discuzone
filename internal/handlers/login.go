package handlers

import (
	"database/sql"
	"encoding/json"
	"forum/internal"
	"forum/internal/utils"
	"net/http"
	"strings"
	"text/template"
)

// Function for rendering an HTML form
func ServeFormLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create data struct to pass to template
		data := struct {
			Email    string
			Password string
		}{
			Email:    r.URL.Query().Get("email"),
			Password: r.URL.Query().Get("password"),
		}

		tmpl, err := template.ParseFiles(
			"templates/layout_auth.html",
			"templates/login.html",
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
	}
}

func HandlerLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		// Parse JSON instead of a form
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request data")
			return
		}

		credentials.Email = strings.TrimSpace(credentials.Email)
		if credentials.Email == "" || credentials.Password == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email and password are required")
			return
		}

		oauthMarker := false
		utils.ValidateAndLoginUser(w, r, db, credentials, oauthMarker)

	}
}
