package test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"forum/internal/handlers"
)

func TestUpdateCommentHandler_Success(t *testing.T) {
	db, teardown := SetupTestDB(t)
	defer teardown()

	// 1. Створюємо тестові дані
	_, err := db.Exec(`INSERT INTO users (id, username, role) VALUES (?, ?, ?)`, 1, "testuser", "admin")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO posts (id, title, content) VALUES (?, ?, ?)`, 5, "Test Post", "Post content")
	if err != nil {
		t.Fatalf("failed to insert test post: %v", err)
	}

	_, err = db.Exec(`INSERT INTO comments (id, user_id, post_id, content) VALUES (?, ?, ?, ?)`, 10, 1, 5, "Original comment")
	if err != nil {
		t.Fatalf("failed to insert test comment: %v", err)
	}

	// 2. Створюємо POST-запит
	form := url.Values{}
	form.Add("commentId", "10")
	form.Add("commentContent", "Updated comment")

	req := httptest.NewRequest(http.MethodPost, "/update-comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 3. Додаємо cookie або заголовок для ідентифікації користувача
	// Припустимо, твій хендлер читає cookie "user_id"
	req.AddCookie(&http.Cookie{
		Name:  "user_id",
		Value: "1",
	})

	w := httptest.NewRecorder()

	// 4. Викликаємо хендлер
	handler := handlers.UpdateCommentHandler(db)
	handler.ServeHTTP(w, req)

	// 5. Перевірка статусу
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d", w.Code)
	}

	// 6. Перевірка, що коментар оновився
	var content string
	err = db.QueryRow(`SELECT content FROM comments WHERE id = ?`, 10).Scan(&content)
	if err != nil {
		t.Fatalf("failed to query updated comment: %v", err)
	}

	if content != "Updated comment" {
		t.Errorf("expected comment content to be updated, got %q", content)
	}
}

