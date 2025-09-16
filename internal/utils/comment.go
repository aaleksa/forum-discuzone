package utils

import (
	"database/sql"
	"fmt"
	"log"
	// "forum/internal"
	"forum/internal/models"
	"time"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func GetCommentsCount(db *sql.DB, postID int) (int, error) {
	// Query to count the number of comments
	var commentsCount int
	err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = ?", postID).Scan(&commentsCount)
	if err != nil {
		return 0, err
	}
	return commentsCount, nil
}

// Get all comments on a post
func GetCommentsByPostID(db *sql.DB, postID int) ([]models.Comment, error) {
	// log.Printf("[DEBUG] GetCommentsByPostID called for postID=%d", postID)
	var rawCreatedAt time.Time
	query := `
       SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
    `
	rows, err := db.Query(query, postID)
	if err != nil {
		log.Printf("[ERROR] Failed to query comments for post %d: %v", postID, err)
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.UserName, &comment.Content, &rawCreatedAt); err != nil {
			log.Printf("[ERROR] Failed to scan comment row: %v", err)
			return nil, err
		}
		// Format the date
		comment.CreatedAt = FormatDate(rawCreatedAt)
		// Get like/dislike counts for each comment
		comment.Likes, comment.Dislikes, err = GetCommentReactionsCount(db, comment.ID)
		if err != nil {
			log.Printf("[ERROR] Failed to get reactions for comment %d: %v", comment.ID, err)
			return nil, fmt.Errorf("error getting comment reactions: %w", err)
		}
		// log.Printf("[DEBUG] Comment fetched: ID=%d, UserID=%d, UserName=%s, Likes=%d, Dislikes=%d", comment.ID, comment.UserID, comment.UserName, comment.Likes, comment.Dislikes)

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Rows iteration error for post %d: %v", postID, err)
		return nil, err
	}

	// log.Printf("[DEBUG] Total comments returned for post %d: %d", postID, len(comments))

	return comments, nil
}

// Get only comment texts (useful for a concise view)
func GetCommentTextsByPostID(db *sql.DB, postID int) ([]string, error) {
	query := "SELECT content FROM comments WHERE post_id = ?"
	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var texts []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, err
		}
		texts = append(texts, content)
	}

	return texts, nil
}

func GetCommentByID(db *sql.DB, commentID int) (models.Comment, error) {
	query := `
		SELECT c.id, c.content, c.user_id, u.username, c.post_id, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`

	var comment models.Comment
	row := db.QueryRow(query, commentID)
	err := row.Scan(&comment.ID, &comment.Content, &comment.UserID, &comment.UserName, &comment.PostID, &comment.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Comment{}, sql.ErrNoRows
		}
		return models.Comment{}, err
	}

	return comment, nil
}

// AddComment inserts a new comment and creates a notification for the post author.
func AddComment(db *sql.DB, postID, userID, parentCommentID int, content string) (int, error) {
	post, err := GetPostByID(db, postID)
	if err != nil {
		return 0, fmt.Errorf("failed to get post: %w", err)
	}

	// Insert comment and get its ID
	res, err := db.Exec(
		"INSERT INTO comments (post_id, user_id, parent_comment_id, content, created_at) VALUES (?, ?, ?, ?, datetime('now'))",
		postID, userID, parentCommentID, content,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert comment: %w", err)
	}

	commentID64, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	commentID := int(commentID64)

	notificationType := "comment"
	if parentCommentID != 0 {
		notificationType = "reply"
	}

	if err := CreateNotification(db, post.UserID, userID, postID, commentID, notificationType); err != nil {
		return commentID, fmt.Errorf("comment created but notification failed: %w", err)
	}

	return commentID, nil
}
