package utils

import (
	"database/sql"
)

// CreateNotification inserts a new notification into the DB
func CreateNotification(db *sql.DB, recipientID, actorID, postID, commentID int, notifType string) error {
	if recipientID == actorID {
		return nil // Don't notify yourself
	}
	_, err := db.Exec(`
		INSERT INTO notifications (user_id, actor_id, post_id, comment_id, type)
		VALUES (?, ?, ?, ?, ?)`, recipientID, actorID, postID, commentID, notifType)
	return err
}
