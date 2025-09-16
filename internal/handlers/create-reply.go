package handlers

import (
	"database/sql"
	"encoding/json"
	// "sync"
	// "forum/internal/models"
	"forum/internal/utils"
	"net/http"
	"strconv"
	"time"
	// "github.com/gorilla/websocket"
)

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID int, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// TrimContent shortens content for notifications
func TrimContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "..."
}

func HandlerAddReply(db *sql.DB, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		userID, ok := utils.MustGetUserID(w, r)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if err := r.ParseForm(); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid form")
			return
		}

		content := r.FormValue("reply_content")
		if content == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Reply content cannot be empty")
			return
		}

		parentCommentID, err := strconv.Atoi(r.FormValue("parent_comment_id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid parent comment ID")
			return
		}

		postID, err := strconv.Atoi(r.FormValue("post_id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid post ID")
			return
		}

		// Add comment
		commentID, err := utils.AddComment(db, postID, userID, parentCommentID, content)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not add reply")
			return
		}

		// If this is a reply to another comment
		if parentCommentID != 0 {
			// Get parent comment author info
			var parentAuthorID int
			var parentAuthorName, parentContent string
			err := db.QueryRow(`
				SELECT c.user_id, u.username, c.content
				FROM comments c
				JOIN users u ON c.user_id = u.id
				WHERE c.id = ?`, parentCommentID).Scan(&parentAuthorID, &parentAuthorName, &parentContent)

			if err != nil {
				if err == sql.ErrNoRows {
					utils.RespondWithError(w, http.StatusBadRequest, "Parent comment not found")
					return
				}
				utils.RespondWithError(w, http.StatusInternalServerError, "Database error")
				return
			}

			// Don't notify yourself
			if parentAuthorID != userID {
				// Get post info
				var postTitle string
				err := db.QueryRow("SELECT title FROM posts WHERE id = ?", postID).Scan(&postTitle)
				if err != nil {
					utils.RespondWithError(w, http.StatusInternalServerError, "Could not get post info")
					return
				}

				// Get current user's username
				var currentUsername string
				err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&currentUsername)
				if err != nil {
					utils.RespondWithError(w, http.StatusInternalServerError, "Could not get user info")
					return
				}

				// Prepare WebSocket notification
				notification := map[string]interface{}{
					"type":              "new_reply",
					"post_id":           postID,
					"post_title":        postTitle,
					"comment_id":        commentID,
					"parent_comment_id": parentCommentID,
					"parent_content":    TrimContent(parentContent, 100),
					"actor_id":          userID,
					"actor_name":        currentUsername,
					"content":           TrimContent(content, 100),
					"created_at":        time.Now().Format(time.RFC3339),
					"unread_count":      1,
				}

				// Send WebSocket notification
				message, err := json.Marshal(notification)
				if err != nil {
					utils.RespondWithError(w, http.StatusInternalServerError, "Could not prepare notification")
					return
				}

				if hub != nil {
					hub.BroadcastToUser(parentAuthorID, message)
				}

				// Add notification to database
				if err := utils.CreateNotification(db, parentAuthorID, userID, postID, commentID, "reply"); err != nil {
					utils.RespondWithError(w, http.StatusInternalServerError, "Could not save notification")
					return
				}
			}
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"reply": map[string]interface{}{
				"id":                commentID,
				"parent_comment_id": parentCommentID,
				"post_id":           postID,
				"user_id":           userID,
				"content":           content,
				"created_at":        time.Now().Format("2006-01-02 15:04:05"),
			},
		})
	}
}
