package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/internal/models"
	"forum/internal/utils"
	"log"
	"net/http"
)

// HandlerGetNotifications handles fetching unread notifications for a user
func HandlerGetNotifications(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := utils.MustGetUserID(w, r)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Запит тільки для непрочитаних сповіщень
		rows, err := db.Query(`
            SELECT n.id, n.type, n.post_id, n.actor_id, u.username, n.created_at
            FROM notifications n
            JOIN users u ON n.actor_id = u.id
            WHERE n.user_id = ? AND n.is_read = FALSE
            ORDER BY n.created_at DESC
        `, userID)
		if err != nil {
			log.Printf("Database query error: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}
		defer rows.Close()

		var notifications []map[string]interface{}
		for rows.Next() {
			var id, postID, actorID int
			var notifType, actorUsername, createdAt string

			if err := rows.Scan(&id, &notifType, &postID, &actorID, &actorUsername, &createdAt); err != nil {
				log.Printf("Error scanning notification row: %v", err)
				continue
			}

			// Отримуємо пост для заголовка
			post, err := utils.GetPostByID(db, postID)
			if err != nil {
				log.Printf("Error getting post title: %v", err)
				continue
			}

			notifications = append(notifications, map[string]interface{}{
				"id":         id,
				"type":       notifType,
				"post_id":    postID,
				"post_title": post.Title, // Додаємо заголовок посту
				"actor":      actorUsername,
				"is_read":    false,
				"created_at": createdAt,
			})
		}

		if err := rows.Err(); err != nil {
			log.Printf("Rows iteration error: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Data processing error")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, notifications)
	}
}

// HandlerMarkNotificationRead marks a specific notification as read
func HandlerMarkNotificationRead(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			NotificationID int `json:"notification_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		userID, ok := utils.MustGetUserID(w, r)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		_, err := db.Exec(`
			UPDATE notifications 
			SET is_read = 1 
			WHERE id = ? AND user_id = ?`,
			payload.NotificationID, userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not mark notification as read")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "read"})
	}
}

// HandlerMarkAllNotificationsRead marks all notifications for the current user as read
func HandlerMarkAllNotificationsRead(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := utils.MustGetUserID(w, r)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		_, err := db.Exec(`UPDATE notifications SET is_read = 1 WHERE user_id = ?`, userID)
		if err != nil {
			log.Println("DB Exec error:", err) // ← це виведе справжню помилку
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not mark all notifications as read")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "all_read"})
	}
}

// GetAllNotifications повертає всі сповіщення для користувача
func GetAllNotifications(db *sql.DB, userID int) ([]models.Notification, error) {
	query := `
        SELECT 
            n.id, 
            n.type, 
            n.post_id, 
            p.title as post_title,
            n.actor_id, 
            u.username as actor_name,
            n.is_read,
            n.created_at
        FROM notifications n
        JOIN users u ON n.actor_id = u.id
        JOIN posts p ON n.post_id = p.id
        WHERE n.user_id = ?
        ORDER BY n.created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification

	for rows.Next() {
		var n models.Notification
		err := rows.Scan(
			&n.ID,
			&n.Type,
			&n.PostID,
			&n.PostTitle,
			&n.ActorID,
			&n.ActorName,
			&n.IsRead,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return notifications, nil
}
