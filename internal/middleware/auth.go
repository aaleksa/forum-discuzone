package middleware

import (
	"context"
	"database/sql"
	"forum/internal/security"
	"forum/internal/utils"
	"log"
	"net/http"
	"time"
)

func AuthMiddleware(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := security.ValidateSession(db, r)
		log.Printf("User (ID: %d) successfully AuthMiddleware", userID)

		// Don't redirect if there is no session - this is the norm for guests

		cookie, err := r.Cookie("session_id")
		if err == nil {
			// Get the session expires_at
			var expiresAt time.Time
			err := db.QueryRow("SELECT expires_at FROM sessions WHERE id = ?", cookie.Value).Scan(&expiresAt)
			if err == nil {
				// If less than an hour remains, update the session
				if time.Until(expiresAt) < time.Hour {
					err = security.RefreshSession(db, cookie.Value)
					if err == nil {
						newExpiry := time.Now().Add(24 * time.Hour)
						security.SetSessionCookie(w, cookie.Value, newExpiry, true)
					}
				}
			}
		}

		// Add userID (can be 0 or nil if not logged in)
		ctx := context.WithValue(r.Context(), utils.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
