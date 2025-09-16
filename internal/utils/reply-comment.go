package utils

import (
	"database/sql"
	"forum/internal/models"
)

// Gets all replies to the post's comments
func GetRepliesForComments(db *sql.DB, postID int) (map[int][]models.Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.parent_comment_id, c.created_at, u.username
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ? AND c.parent_comment_id IS NOT NULL
		ORDER BY c.created_at ASC
	`

	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	repliesMap := make(map[int][]models.Comment)

	for rows.Next() {
		var c models.Comment
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.ParentCommentID, &c.CreatedAt, &c.Username)
		if err != nil {
			return nil, err
		}
		repliesMap[c.ParentCommentID] = append(repliesMap[c.ParentCommentID], c)
	}

	return repliesMap, nil
}
