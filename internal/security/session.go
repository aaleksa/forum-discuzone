package security

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var isProduction = os.Getenv("APP_ENV") == "production" // automatic environment detection

func CreateSession(w http.ResponseWriter, userID int, db *sql.DB) error {
	log.Printf("Processing CreateSession, userID: %d:", userID)
	const maxRetries = 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * 100 * time.Millisecond)
		}

		err := tryCreateSession(w, userID, db)
		if err == nil {
			return nil
		}

		if !strings.Contains(err.Error(), "database is locked") {
			return err
		}
		lastErr = err
	}

	return fmt.Errorf("after %d attempts: %v", maxRetries, lastErr)
}

func tryCreateSession(w http.ResponseWriter, userID int, db *sql.DB) error {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Use transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// First delete old sessions
	if _, err := tx.Exec("DELETE FROM sessions WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("failed to delete old sessions: %v", err)
	}

	// Then add a new session
	if _, err := tx.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expiresAt,
	); err != nil {
		return fmt.Errorf("failed to insert new session: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	SetSessionCookie(w, sessionID, expiresAt, isProduction)
	return nil
}

func ValidateSession(db *sql.DB, r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_id")

	if err != nil {
		log.Printf("ValidateSession: no session cookie found")
		return 0, fmt.Errorf("session cookie not found")
	}
	log.Printf("ValidateSession: cookie value: %s", cookie.Value)

	if err != nil {
		return 0, fmt.Errorf("session cookie not found")
	}

	sessionID, ok := VerifySignedSessionID(cookie.Value)
	if !ok {
		return 0, fmt.Errorf("invalid cookie signature")
	}
	log.Printf("Processing ValidateSession, session_id: %s,", sessionID)
	var userID int
	var expiresAt time.Time
	err = db.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("invalid session")
		}
		return 0, fmt.Errorf("database error: %v", err)
	}

	if time.Now().After(expiresAt) {
		_, _ = db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		return 0, fmt.Errorf("session expired")
	}

	return userID, nil
}

func DestroySession(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return err
	}

	sessionID, ok := VerifySignedSessionID(cookie.Value)
	if !ok {
		return fmt.Errorf("invalid cookie signature")
	}

	_, err = db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return err
	}

	// Delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

func SetSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time, secure bool) {
	signed := SignSessionID(sessionID)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  expiresAt,
	})
}

func RefreshSession(db *sql.DB, sessionID string) error {
	newExpiry := time.Now().Add(24 * time.Hour)
	_, err := db.Exec("UPDATE sessions SET expires_at = ? WHERE id = ?", newExpiry, sessionID)
	return err
}
