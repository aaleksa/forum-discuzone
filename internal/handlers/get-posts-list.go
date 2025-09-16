package handlers

import (
	"database/sql"
	"forum/internal/models"
	"forum/internal/utils"
	"time"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func GetPosts(db *sql.DB, currentUser *models.User) ([]models.PostView, error) {
	rows, err := db.Query(`
        SELECT p.id, p.user_id, p.title, p.content, p.created_at, u.username
        FROM posts p
        JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostView
	for rows.Next() {
		var post models.PostView
		var rawCreatedAt time.Time

		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &rawCreatedAt, &post.UserName); err != nil {
			return nil, err
		}

		// Format the date
		post.CreatedAt = utils.FormatDate(rawCreatedAt)

		// Getting the number of likes
		post.Likes, post.Dislikes, err = utils.GetPostLikesCount(db, post.ID)
		if err != nil {
			return nil, err
		}
		post.CommentsCount, err = utils.GetCommentsCount(db, post.ID)
		if err != nil {
			return nil, err
		}
		// Receiving comments
		post.Comments, err = utils.GetCommentsByPostID(db, post.ID)
		if err != nil {
			return nil, err
		}
		// Receiving tags
		post.Tags, err = utils.GetPostTags(db, post.ID)
		if err != nil {
			return nil, err
		}

		// Receiving images
		post.ImagePaths, err = utils.GetPostImages(db, post.ID)
		if err != nil {
			return nil, err
		}

		// Store current user info with each post
		post.CurrentUser = currentUser

		posts = append(posts, post)
	}

	return posts, nil
}
