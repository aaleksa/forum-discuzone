package handlers

import (
	"database/sql"
	"forum/internal"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func HandlerDeletePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the post ID
		postID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/delete_post/"))
		if err != nil {
			log.Printf("Invalid post ID: %v", err)
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid post ID.")
			return
		}

		// // Get current user from session
		currentUser, err := utils.GetUserFromSession(w, r, db)
		if err != nil || currentUser == nil {
			log.Printf("Unauthorized delete attempt: %v", err)
			errors.RenderError(w, http.StatusUnauthorized, "Unauthorized", "Please log in.")
			return
		}

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Transaction begin error: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Database error.")
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
			}
		}()

		// 1. Check post exists and get author ID
		var authorID int
		err = tx.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&authorID)
		if err != nil {
			if err == sql.ErrNoRows {
				errors.RenderError(w, http.StatusNotFound, "Not Found", "Post not found.")
			} else {
				errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Database error.")
			}
			return
		}

		// 2. Check permissions (admin, moderator, or post author)
		if !utils.HasPermission(currentUser, authorID, "edit") {
			errors.RenderError(w, http.StatusForbidden, "Forbidden", "You don't have permission to edit this post.")
			return
		}

		// 3. Delete related data first (foreign key constraints)
		_, err = tx.Exec("DELETE FROM likes WHERE post_id = ?", postID)
		if err != nil {
			log.Printf("Error deleting likes: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Delete failed.")
			return
		}

		_, err = tx.Exec("DELETE FROM comments WHERE post_id = ?", postID)
		if err != nil {
			log.Printf("Error deleting comments: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Delete failed.")
			return
		}

		// 4. Delete the post
		res, err := tx.Exec("DELETE FROM posts WHERE id = ?", postID)
		if err != nil {
			log.Printf("Delete error: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Delete failed.")
			return
		}

		// 5. Verify deletion
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			errors.RenderError(w, http.StatusNotFound, "Not Found", "Post not found.")
			return
		}

		// 6. Commit transaction
		if err := tx.Commit(); err != nil {
			log.Printf("Commit error: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Delete failed.")
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
