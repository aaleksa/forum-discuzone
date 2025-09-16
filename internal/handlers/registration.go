package handlers

import (
	"database/sql"
	"encoding/json"
	"forum/internal/models"
	"forum/internal/security"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"net/http"
	"strings"
	"text/template"
)

// ServeFormRegister renders the registration form template with empty fields
func ServeFormRegister(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(
			"templates/layout_auth.html",
			"templates/registration.html",
			"templates/header_auth.html",
		)
		if err != nil {
			http.Error(w, "Internal Server Error: Template not loaded", http.StatusInternalServerError)
			return
		}

		emptyForm := models.RegisterPageData{
			Username: "",
			Email:    "",
			Password: "",
			Error:    "",
		}

		err = tmpl.ExecuteTemplate(w, "layout", emptyForm)
		if err != nil {
			http.Error(w, "Internal Server Error: Template execution error", http.StatusInternalServerError)
		}
	}
}

// HandlerRegistration handles user registration requests
func HandlerRegistration(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set content type for JSON responses
		w.Header().Set("Content-Type", "application/json")

		// Only accept POST requests
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Method Not Allowed",
			})
			return
		}

		// Parse form data
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid form data",
			})
			return
		}

		// Get and trim form values
		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")
		provider := r.FormValue("provider")      // e.g. "google"
		providerID := r.FormValue("provider_id") // e.g. Google's sub claim

		// Validate inputs
		if username == "" || email == "" || password == "" || confirmPassword == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "All fields are required",
			})
			return
		}

		if password != confirmPassword {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Passwords do not match",
			})
			return
		}

		if len(password) < 8 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Password must be at least 8000 characters",
			})
			return
		}

		// Hash password
		hashedPassword, err := security.HashPassword(password)
		if err != nil {
			log.Println("Password hashing error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			})
			return
		}
		// Check if user exists
		if IsUserExists(db, username, email, provider, providerID) {
			var errorMsg string
			if provider != "" {
				errorMsg = "User with this email or social account already exists"
			} else {
				errorMsg = "Username or email already exists"
			}

			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error": errorMsg,
			})
			return
		}

		// Create user
		if err := СreateRegularUser(db, username, email, hashedPassword); err != nil {
			log.Println("Database error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to register user",
			})
			return
		}

		// Success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Registration successful",
		})
	}
}
