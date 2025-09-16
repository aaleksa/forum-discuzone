package test

import (
	"database/sql"
	"testing"

	"forum/internal/handlers"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// func openTestDB(t *testing.T) *sql.DB {
// 	db, err := sql.Open("sqlite3", ":memory:")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// Створюємо потрібну таблицю likes (підлаштуй під свою структуру)
// 	_, err = db.Exec(`
// 	CREATE TABLE IF NOT EXISTS likes (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		user_id INTEGER NOT NULL,
// 		post_id INTEGER,
// 		comment_id INTEGER,
// 		reaction TEXT NOT NULL CHECK (reaction IN ('Like', 'Dislike')),
// 		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
// 		CHECK (post_id IS NOT NULL OR comment_id IS NOT NULL),
// 		UNIQUE(user_id, post_id, comment_id),
// 		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
// 		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
// 		FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
// 	);`)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	return db
// }

func TestProcessReaction(t *testing.T) {
	db, teardown := SetupTestDB(t)
	defer teardown()

	// Створюємо пост від іншого користувача
	res, err := db.Exec("INSERT INTO posts (user_id, title, content, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)", 2, "Test post", "Content")
	if err != nil {
		t.Fatalf("failed to insert post: %v", err)
	}
	postID, _ := res.LastInsertId()
	userID := 1 // користувач, який ставить реакцію

	// Додаємо лайк
	err = handlers.ProcessReactionForPost(db, userID, int(postID), "like")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Перевіряємо, що лайк додано
	var reaction string
	err = db.QueryRow("SELECT reaction FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&reaction)
	if err != nil {
		t.Fatalf("expected row, got error %v", err)
	}
	if reaction != "Like" {
		t.Errorf("expected reaction 'Like', got '%s'", reaction)
	}

	// Змінюємо реакцію на dislike
	err = handlers.ProcessReactionForPost(db, userID, int(postID), "dislike")
	if err != nil {
		t.Fatalf("expected no error on reaction change, got %v", err)
	}
	err = db.QueryRow("SELECT reaction FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&reaction)
	if err != nil {
		t.Fatalf("expected row after update, got error %v", err)
	}
	if reaction != "Dislike" {
		t.Errorf("expected reaction 'Dislike', got '%s'", reaction)
	}

	// Вилучаємо реакцію
	err = handlers.ProcessReactionForPost(db, userID, int(postID), "dislike")
	if err != nil {
		t.Fatalf("expected no error on reaction delete, got %v", err)
	}
	err = db.QueryRow("SELECT reaction FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&reaction)
	if err != sql.ErrNoRows {
		t.Errorf("expected no rows after delete, got %v", err)
	}
}

func TestProcessCommentReaction(t *testing.T) {
	db, teardown := SetupTestDB(t)
	defer teardown()

	userID, commentID := 1, 20

	// Додаємо лайк до коментаря
	err := handlers.ProcessReactionForComment(db, userID, commentID, "like")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Перевіряємо, що лайк додано
	var reaction string
	err = db.QueryRow("SELECT reaction FROM likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&reaction)
	if err != nil {
		t.Fatalf("expected row, got error %v", err)
	}
	if reaction != "Like" {
		t.Errorf("expected reaction 'Like', got '%s'", reaction)
	}

	// Змінюємо реакцію на dislike
	err = handlers.ProcessReactionForComment(db, userID, commentID, "dislike")
	if err != nil {
		t.Fatalf("expected no error on reaction change, got %v", err)
	}
	err = db.QueryRow("SELECT reaction FROM likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&reaction)
	if err != nil {
		t.Fatalf("expected row after update, got error %v", err)
	}
	if reaction != "Dislike" {
		t.Errorf("expected reaction 'Dislike', got '%s'", reaction)
	}

	// Вилучаємо реакцію (клік на ту ж реакцію видаляє її)
	err = handlers.ProcessReactionForComment(db, userID, commentID, "dislike")
	if err != nil {
		t.Fatalf("expected no error on reaction delete, got %v", err)
	}
	err = db.QueryRow("SELECT reaction FROM likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&reaction)
	if err != sql.ErrNoRows {
		t.Errorf("expected no rows after delete, got %v", err)
	}
}
