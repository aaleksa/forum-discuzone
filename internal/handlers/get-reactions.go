package handlers

import (
	"database/sql"
	"fmt"
	"strings"
)

// getUserReaction returns the user's reaction (like/dislike) for a given content type and ID
func getUserReaction(db *sql.DB, userID int, contentType string, contentID int) (string, error) {
	var reaction string
	var query string
	var err error

	switch contentType {
	case "post":
		query = `
			SELECT reaction 
			FROM likes 
			WHERE user_id = ? AND post_id = ?`
		err = db.QueryRow(query, userID, contentID).Scan(&reaction)

	case "comment":
		query = `
			SELECT reaction 
			FROM likes 
			WHERE user_id = ? AND comment_id = ?`
		err = db.QueryRow(query, userID, contentID).Scan(&reaction)

	default:
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // no reaction
		}
		return "", err
	}

	return strings.ToLower(reaction), nil
}
