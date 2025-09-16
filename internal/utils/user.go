package utils

import (
	"database/sql"
	"time"
)

func FindOrCreateUserByEmail(email string, db *sql.DB) (int, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Створити нового користувача
		res, err := db.Exec("INSERT INTO users (email, created_at) VALUES (?, ?)", email, time.Now())
		if err != nil {
			return 0, err
		}
		lastID, _ := res.LastInsertId()
		return int(lastID), nil
	} else if err != nil {
		return 0, err
	}
	return userID, nil
}
