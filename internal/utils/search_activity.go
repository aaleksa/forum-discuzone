package utils

import (
	"database/sql"
	"fmt"
	"forum/internal/models"
	"strings"
	"time"
)

func SearchUserActivity(db *sql.DB, userID int, query, activityType string) (models.UserActivityResults, error) {
	query = strings.ToLower(query)
	likeQuery := "%" + query + "%"
	var results models.UserActivityResults

	// Search posts
	if activityType == "" || activityType == "post" {
		rows, err := db.Query(`
            SELECT p.id, p.title, p.content, p.created_at, u.username
            FROM posts p
            JOIN users u ON p.user_id = u.id
            WHERE p.user_id = ? AND (LOWER(p.title) LIKE ? OR LOWER(p.content) LIKE ?)
        `, userID, likeQuery, likeQuery)
		if err != nil {
			return results, fmt.Errorf("error searching posts: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var p models.PostView
			var createdAt time.Time
			if err := rows.Scan(&p.ID, &p.Title, &p.Content, &createdAt, &p.UserName); err != nil {
				return results, fmt.Errorf("error scanning post: %w", err)
			}
			p.CreatedAt = FormatDate(createdAt)
			results.Posts = append(results.Posts, p)
		}
		if err := rows.Err(); err != nil {
			return results, fmt.Errorf("post rows error: %w", err)
		}
	}

	// Search comments
	if activityType == "" || activityType == "comment" {
		rows, err := db.Query(`
            SELECT 
                c.id,
                c.post_id,
                c.user_id,
                c.content,
                0 AS likes,
                0 AS dislikes,
                c.created_at,
                COALESCE(c.parent_comment_id, 0) AS parent_comment_id,
                u.username,
                p.title AS post_title
            FROM comments c
            JOIN posts p ON p.id = c.post_id
            JOIN users u ON u.id = c.user_id
            WHERE c.user_id = ? AND LOWER(c.content) LIKE ?
        `, userID, likeQuery)
		if err != nil {
			return results, fmt.Errorf("error searching comments: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var c models.Comment
			var createdAt time.Time
			if err := rows.Scan(
				&c.ID,
				&c.PostID,
				&c.UserID,
				&c.Content,
				&c.Likes,
				&c.Dislikes,
				&createdAt,
				&c.ParentCommentID,
				&c.Username,
				&c.PostTitle,
			); err != nil {
				return results, fmt.Errorf("error scanning comment: %w", err)
			}
			c.CreatedAt = FormatDate(createdAt)
			results.Comments = append(results.Comments, c)
		}
		if err := rows.Err(); err != nil {
			return results, fmt.Errorf("comment rows error: %w", err)
		}
	}

	// Search likes
	if activityType == "" || activityType == "like" {
		rows, err := db.Query(`
            SELECT p.id, p.title, p.content, p.created_at, u.username
            FROM likes l
            JOIN posts p ON p.id = l.post_id
            JOIN users u ON p.user_id = u.id
            WHERE l.user_id = ? AND (LOWER(p.title) LIKE ? OR LOWER(p.content) LIKE ?)
        `, userID, likeQuery, likeQuery)
		if err != nil {
			return results, fmt.Errorf("error searching likes: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var p models.PostView
			var createdAt time.Time
			if err := rows.Scan(&p.ID, &p.Title, &p.Content, &createdAt, &p.UserName); err != nil {
				return results, fmt.Errorf("error scanning like: %w", err)
			}
			p.CreatedAt = FormatDate(createdAt)
			results.Likes = append(results.Likes, p)
		}
		if err := rows.Err(); err != nil {
			return results, fmt.Errorf("like rows error: %w", err)
		}
	}

	return results, nil
}
