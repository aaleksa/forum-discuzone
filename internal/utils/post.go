package utils

import (
	"context"
	"database/sql"
	"fmt"
	"forum/internal/models"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"strconv"
	"strings"
	"time"
)

func GetPostByID(db *sql.DB, id int) (models.PostView, error) {
	var post models.PostView
	var rawCreatedAt time.Time
	var rawUpdateAt *time.Time

	// Get basic post data + username + image
	row := db.QueryRow(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, u.username,  p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?`, id)

	// Read the data
	if err := row.Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&rawCreatedAt,
		&post.UserName,
		&rawUpdateAt,
	); err != nil {
		return post, err
	}

	// Format the date
	post.CreatedAt = FormatDate(rawCreatedAt)
	// Format the date
	if rawUpdateAt != nil {
		post.UpdatedAt = FormatDate(*rawUpdateAt)
	} else {
		post.UpdatedAt = ""
	}

	post.IsEdited = false
	if post.UpdatedAt != "" && post.UpdatedAt != post.CreatedAt {
		post.IsEdited = true
	}

	// Get the number of likes/dislikes
	var err error
	post.Likes, post.Dislikes, err = GetPostLikesCount(db, post.ID)
	if err != nil {
		return post, err
	}

	// Get the number of comments
	post.CommentsCount, err = GetCommentsCount(db, post.ID)
	if err != nil {
		return post, err
	}

	// Get comments
	post.Comments, err = GetCommentsByPostID(db, post.ID)
	if err != nil {
		return post, err
	}
	// Receiving images
	post.ImagePaths, err = GetPostImages(db, post.ID)
	if err != nil {
		return post, err
	}

	return post, nil
}

func UpdatePostFull(
	ctx context.Context,
	db *sql.DB,
	postID int,
	title, content string,
	categories []string,
	tags string,
	newImagePaths []string,
	removeImageIDs []int,
) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// === Update title and content ===
	if _, err := tx.ExecContext(ctx, `
	    UPDATE posts 
	    SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP
	    WHERE id = ?`,
		title, content, postID); err != nil {
		return fmt.Errorf("update title/content: %w", err)
	}

	// === Deleting selected images ===
	for _, imgID := range removeImageIDs {
		if imgID == 0 {
			continue
		}
		log.Printf("Deleting image ID: %d", imgID)
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM post_images
			WHERE id = ? AND post_id = ?`,
			imgID, postID); err != nil {
			return fmt.Errorf("delete image ID %d: %w", imgID, err)
		}
	}

	// === Adding new images ===
	for i, path := range newImagePaths {
		if path == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO post_images (post_id, image_path, order_index)
			VALUES (?, ?, ?)`,
			postID, path, i); err != nil {
			return fmt.Errorf("insert image '%s': %w", path, err)
		}
	}

	// === Update categories ===
	// Delete old categories
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM post_categories
		WHERE post_id = ?`, postID); err != nil {
		return fmt.Errorf("delete old categories: %w", err)
	}

	// Add new categories
	for _, catIDStr := range categories {
		catID, err := strconv.Atoi(catIDStr)
		if err != nil {
			return fmt.Errorf("invalid category ID '%s': %w", catIDStr, err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO post_categories (post_id, category_id)
			VALUES (?, ?)`, postID, catID); err != nil {
			return fmt.Errorf("insert category %d: %w", catID, err)
		}
	}

	// === Tag Processing ===
	tagList := []string{}
	if tags != "" {
		for _, tag := range strings.Split(tags, ",") {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				tagList = append(tagList, trimmed)
			}
		}
	}

	if err := ProcessPostTags(ctx, tx, int64(postID), tagList); err != nil {
		return fmt.Errorf("processing post tags: %w", err)
	}

	// === Commit all changes ===
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// utils/posts.go
func CheckIfPostLiked(db *sql.DB, postID, userID int) (bool, error) {
	var exists bool
	err := db.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM likes WHERE post_id = ? AND user_id = ?)
    `, postID, userID).Scan(&exists)
	return exists, err
}

func FormatDate(t time.Time) string {
	return t.Format("02 Jan 2006 15:04")
}

func GetCreatedPosts(db *sql.DB, userID int) ([]models.Post, error) {
	var posts []models.Post

	// First, we get the basic data of the posts
	query := "SELECT id, title, content, created_at FROM posts WHERE user_id = ?"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user posts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post data: %w", err)
		}

		// We get the number of likes/dislikes for a post
		likes, dislikes, err := GetPostLikesCount(db, post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get likes count for post %d: %w", post.ID, err)
		}

		post.Likes = likes
		post.Dislikes = dislikes
		post.UserID = userID

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %w", err)
	}

	return posts, nil
}
