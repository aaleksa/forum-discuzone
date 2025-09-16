package handlers

import (
	"database/sql"
	"forum/internal/security"
	"log"
	"net/http"
	"os"
)

var isProduction = os.Getenv("APP_ENV") == "production" // automatic environment detection

// Function to delete the session cookie
func LogoutHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Call DestroySession for the main exit logic
		if err := security.DestroySession(w, r, db); err != nil {
			log.Printf("Failed to destroy session: %v", err)
			// Continue the process even if there is an error
		}

		// 2. Additionally delete other cookies (if they exist)
		removeCookie := func(name string) {
			http.SetCookie(w, &http.Cookie{
				Name:     name,
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   isProduction,
				SameSite: http.SameSiteStrictMode,
			})
		}

		removeCookie("user_id")    // Delete additional cookies
		removeCookie("csrf_token") // Example of another cookie

		// 3. Додаткові заходи безпеки
		w.Header().Add("Clear-Site-Data", `"cookies"`) // Modern cleaning method

		// 4. Redirect to the main page
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
