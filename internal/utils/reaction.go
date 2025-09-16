package utils

import (
	"database/sql"
	"fmt"
	"forum/internal/models"
	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// getLikesCount returns the number of likes and dislikes for a post
func GetPostLikesCount(db *sql.DB, postID int) (likes int, dislikes int, err error) {
	query := `
        SELECT
            COALESCE(SUM(CASE WHEN reaction = 'Like' THEN 1 ELSE 0 END), 0) AS likes,
            COALESCE(SUM(CASE WHEN reaction = 'Dislike' THEN 1 ELSE 0 END), 0) AS dislikes
        FROM likes
        WHERE post_id = ?`
	err = db.QueryRow(query, postID).Scan(&likes, &dislikes)
	if err != nil {
		return 0, 0, err
	}
	return likes, dislikes, nil
}

// getCommentReactionsCount returns the number of likes and dislikes for a comment
func GetCommentReactionsCount(db *sql.DB, commentID int) (likes int, dislikes int, err error) {
	query := `
        SELECT
            COALESCE(SUM(CASE WHEN reaction = 'Like' THEN 1 ELSE 0 END), 0) AS likes,
            COALESCE(SUM(CASE WHEN reaction = 'Dislike' THEN 1 ELSE 0 END), 0) AS dislikes
        FROM likes
        WHERE comment_id = ?`
	err = db.QueryRow(query, commentID).Scan(&likes, &dislikes)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get comment reactions: %w", err)
	}
	return likes, dislikes, nil
}

func GetLikedPosts(db *sql.DB, userID int) ([]models.LikedPosts, error) {
	var posts []models.LikedPosts
	query := `
       SELECT p.id, p.title
        FROM posts p
        JOIN likes l ON p.id = l.post_id
        WHERE l.user_id = ? AND l.reaction = 'Like'`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.LikedPosts
		if err := rows.Scan(&post.ID, &post.Title); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %v", err)
	}
	return posts, nil
}

func GetDislikes(db *sql.DB, userID int) ([]models.DislikePosts, error) {
	var posts []models.DislikePosts
	query := `
        SELECT p.id, p.title
        FROM posts p
        JOIN likes l ON p.id = l.post_id
        WHERE l.user_id = ? AND l.reaction = 'Dislike'`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.DislikePosts
		if err := rows.Scan(&post.ID, &post.Title); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %v", err)
	}
	return posts, nil
}
