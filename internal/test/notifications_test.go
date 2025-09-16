package test

import (
	"forum/internal/models"
	"forum/internal/utils"
	"testing"
)

func TestCreateNotification(t *testing.T) {
	db, teardown := SetupTestDB(t)
	defer teardown()

	// Створюємо нове сповіщення
	err := utils.CreateNotification(db, 1, 2, 1, 1, "comment")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Перевірка: сповіщення реально додалося
	var n models.Notification
	err = db.QueryRow(`SELECT type, post_id, actor_id 
	                   FROM notifications 
	                   WHERE post_id=? AND actor_id=? AND type=?`,
		1, 2, "comment").Scan(&n.Type, &n.PostID, &n.ActorID)
	if err != nil {
		t.Fatalf("Notification was not inserted: %v", err)
	}

	if n.Type != "comment" || n.PostID != 1 || n.ActorID != 2 {
		t.Errorf("Inserted notification has wrong values: %+v", n)
	}
}
