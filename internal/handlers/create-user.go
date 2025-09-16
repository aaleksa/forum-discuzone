package handlers

import (
	"database/sql"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"time"
)

// Creates a regular user (with password)
func СreateRegularUser(db *sql.DB, username, email, hashedPassword string) error {
	_, err := db.Exec(`
        INSERT INTO users (username, email, password, created_at)
        VALUES (?, ?, ?, ?)`,
		username, email, hashedPassword, time.Now())
	return err
}

// Creates an OAuth user (without password)
func СreateOAuthUser(db *sql.DB, username, email, provider, providerID, avatarURL string) error {
	_, err := db.Exec(`
        INSERT INTO users (username, email, password, provider, provider_id, avatar_url, created_at)
        VALUES (?, ?, '', ?, ?, ?, ?)`,
		username, email, provider, providerID, avatarURL, time.Now())
	return err
}

// isUserExists checks if a user with the given username or email already exists
func IsUserExists(db *sql.DB, username, email, provider, providerID string) bool {
	var count int
	var err error

	// Якщо це OAuth реєстрація
	if provider != "" && providerID != "" {
		err = db.QueryRow(`
            SELECT COUNT(*) FROM users 
            WHERE email = ? OR 
                  (provider = ? AND provider_id = ?)`,
			email, provider, providerID).Scan(&count)
	} else {
		// Звичайна реєстрація
		err = db.QueryRow(`
            SELECT COUNT(*) FROM users 
            WHERE username = ? OR email = ?`,
			username, email).Scan(&count)
	}

	if err != nil {
		log.Printf("Database error in isUserExists: %v", err)
		return true // Безпечний варіант - вважати, що користувач існує
	}
	return count > 0
}
