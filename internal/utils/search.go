package utils

import (
	"database/sql"
	"forum/internal/models"
	"strings"
	"time"
)

func SearchPosts(db *sql.DB, query string) ([]models.PostView, error) {
	query = strings.ToLower(query)
	likeQuery := "%" + query + "%"

	rows, err := db.Query(`
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE LOWER(p.title) LIKE LOWER(?)
		   OR LOWER(p.content) LIKE LOWER(?)
	`, likeQuery, likeQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostView

	for rows.Next() {
		var post models.PostView
		var rawCreatedAt time.Time

		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.UserName,
			&post.Title,
			&post.Content,
			&rawCreatedAt,
		); err != nil {
			return nil, err
		}

		post.CreatedAt = FormatDate(rawCreatedAt)

		post.Likes, post.Dislikes, err = GetPostLikesCount(db, post.ID)
		if err != nil {
			return nil, err
		}

		post.CommentsCount, err = GetCommentsCount(db, post.ID)
		if err != nil {
			return nil, err
		}

		post.Comments, err = GetCommentsByPostID(db, post.ID)
		if err != nil {
			return nil, err
		}

		post.Tags, err = GetPostTags(db, post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}
