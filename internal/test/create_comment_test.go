package test

import (
	"context"
	"database/sql"
	"fmt"
	"forum/internal"
	"forum/internal/handlers"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// Mock functions
func getTemplatePath() string {
	// Find the absolute path to the root of the project
	_, b, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(b), "..", "..", "templates")
	return filepath.Join(root, "error.html")
}

func mockOpenDatabase(_ http.ResponseWriter) *sql.DB {
	db, _ := sql.Open("sqlite3", ":memory:")
	db.Exec(`CREATE TABLE comments (id INTEGER PRIMARY KEY, post_id INTEGER, user_id INTEGER, content TEXT)`)
	return db
}

func mockGetUserIDFromSession(_ http.ResponseWriter, _ *http.Request) (int, error) {
	return 1, nil
}

func mockAddComment(_ *sql.DB, postID, userID int, content string) error {
	if content == "fail" {
		return sql.ErrConnDone
	}
	return nil
}

func TestHandlerCreateComment_Success(t *testing.T) {
	errors.Init(getTemplatePath())
	db, teardown := SetupTestDB(t)
	defer teardown()
	handler := handlers.CreateCommentHandler(db)

	// Вставляємо тестовий пост
	res, err := db.Exec("INSERT INTO posts (user_id, title, content, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)", 1, "Hello World", "This is a test post")
	if err != nil {
		t.Fatal(err)
	}
	postID, _ := res.LastInsertId() // отримуємо id вставленого поста

	form := url.Values{}
	form.Add("postId", fmt.Sprint(postID))
	form.Add("content", "Nice post!")

	req := httptest.NewRequest(http.MethodPost, "/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	type contextKey string
	const userIDKey contextKey = "userID"

	// Додаємо userID у контекст
	ctx := context.WithValue(req.Context(), userIDKey, 1)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestHandlerCreateComment_EmptyContent(t *testing.T) {
	errors.Init(getTemplatePath())

	db, teardown := SetupTestDB(t)
	defer teardown()
	handler := handlers.CreateCommentHandler(db)

	form := url.Values{}
	form.Add("postId", "1")
	form.Add("content", "")

	req := httptest.NewRequest(http.MethodPost, "/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d for empty comment, got %d", http.StatusOK, rr.Code)
	}
}

func TestHandlerCreateComment_NotPostMethod(t *testing.T) {
	errors.Init(getTemplatePath())

	db := mockOpenDatabase(nil) // ← fix: create mock DB
	handler := handlers.CreateCommentHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/comment", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
