package handlers

import (
	"database/sql"
	"fmt"
	"forum/internal"
	"forum/internal/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Reaction processing (adapted to your table)
func ProcessReactionForPost(db *sql.DB, userID int, postID int, newReaction string) error {
	// Convert the response to your database format (with uppercase letters)
	dbReaction := strings.Title(strings.ToLower(newReaction)) // "like" -> "Like"

	// Checking the current user reaction
	currentReaction, err := getUserReactionFromDB(db, userID, postID)
	if err != nil {
		return err
	}

	if currentReaction == dbReaction {
		// The user clicked the same button - we delete the reaction
		_, err = db.Exec(`
			DELETE FROM likes 
			WHERE user_id = ? AND post_id = ?`,
			userID, postID)
	} else if currentReaction == "" {
		// New reaction
		_, err = db.Exec(`
			INSERT INTO likes (user_id, post_id, reaction) 
			VALUES (?, ?, ?)`,
			userID, postID, dbReaction)
	} else {
		// Change in reaction
		_, err = db.Exec(`
			UPDATE likes 
			SET reaction = ? 
			WHERE user_id = ? AND post_id = ?`,
			dbReaction, userID, postID)
	}

	if err != nil {
		return err
	}
	// Get the post data
	post, err := utils.GetPostByID(db, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("post with ID %d not found", postID)
		}
		return fmt.Errorf("error getting post: %w", err)
	}

	// Check if this is a reaction to your own post
	if post.UserID == userID {
		log.Printf("User %d is trying to respond to their own post %d", userID, postID)
		return nil
	}

	parentCommentID := 0 // or other default value for the root comment

	if err := utils.CreateNotification(db, post.UserID, userID, postID, parentCommentID, newReaction); err != nil {
		log.Printf("Failed to create notification - From: %d, To: %d, Post: %d, Type: %s. Error: %v",
			userID, post.UserID, postID, newReaction, err)
		// Continue without returning error since notification failure shouldn't block reaction
	}

	return nil
}

// Helper function for getting the response in the original DB format
func getUserReactionFromDB(db *sql.DB, userID int, postID int) (string, error) {
	var reaction string
	query := `
		SELECT reaction 
		FROM likes 
		WHERE user_id = ? AND post_id = ?`

	err := db.QueryRow(query, userID, postID).Scan(&reaction)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return reaction, nil
}

// Reaction handler (customized)
func HandleReaction(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errors.RenderError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "The HTTP method is not supported.")
			return
		}

		// Get the current user
		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Session error.")
			return
		}
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if user == nil || user.ID == 0 {
			errors.RenderError(w, http.StatusUnauthorized, "Error", "User not identified")
			return
		}

		// Getting data from the form
		contentType := r.FormValue("content_type")
		contentIDStr := r.FormValue("content_id")
		reaction := r.FormValue("reaction")

		// Validation
		if contentType == "" || contentIDStr == "" || reaction == "" {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Missing required fields.")
			return
		}

		contentID, err := strconv.Atoi(contentIDStr)
		if err != nil {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid content ID.")
			return
		}

		if reaction != "like" && reaction != "dislike" {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid reaction type.")
			return
		}
		log.Printf("User %d reaction == contentType %d", user.ID, contentType)
		if contentType == "post" {
			err = ProcessReactionForPost(db, user.ID, contentID, reaction)
		} else if contentType == "comment" {
			err = ProcessReactionForComment(db, user.ID, contentID, reaction)
		} else {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid content type.")
			return
		}

		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to process reaction")
			return
		}

		//Redirecting back
		referer := r.Header.Get("Referer")
		if referer == "" {
			referer = "/" // Fallback
		}
		http.Redirect(w, r, referer, http.StatusFound)
	}
}

// Function for handling reactions to comments
func ProcessReactionForComment(db *sql.DB, userID int, commentID int, newReaction string) error {
	dbReaction := strings.Title(strings.ToLower(newReaction))
	log.Printf("ProcessReactionForComment %d", dbReaction)

	//Checking the user's current reaction to the comment
	var currentReaction string
	err := db.QueryRow("SELECT reaction FROM likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&currentReaction)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		currentReaction = ""
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if currentReaction == dbReaction {
		// Removing the reaction
		_, err = tx.Exec("DELETE FROM likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
	} else if currentReaction == "" {
		// Adding a new reaction
		_, err = tx.Exec("INSERT INTO likes (user_id, comment_id, reaction) VALUES (?, ?, ?)", userID, commentID, dbReaction)
	} else {
		// Changing the reaction
		_, err = tx.Exec("UPDATE likes SET reaction = ? WHERE user_id = ? AND comment_id = ?", dbReaction, userID, commentID)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}
