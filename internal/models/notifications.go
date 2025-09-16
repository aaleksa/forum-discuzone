package models

import (
	"time"
)

// Notification структура для повернення даних
type Notification struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	PostID    int       `json:"post_id"`
	PostTitle string    `json:"post_title"`
	ActorID   int       `json:"actor_id"`
	ActorName string    `json:"actor_name"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
