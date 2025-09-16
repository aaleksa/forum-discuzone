package test

import (
	"forum/internal/handlers"
	"testing"
	"time"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func TestCreatePost_Success(t *testing.T) {
	db, teardown := SetupTestDB(t)
	defer teardown()

	// Створимо категорії
	db.Exec("INSERT INTO categories (id, name) VALUES (1, 'Go'), (2, 'Web')")

	userID := 1
	title := "Test Post"
	content := "Some content"
	createdAt := time.Now()
	categoryIDs := []int{1, 2}
	tags := []string{"test", "golang"}
	imagePath := []string{"test1.jpg", "test2.jpg"}
	primaryImageIndex := 1

	err := handlers.CreatePost(db, userID, title, content, createdAt, categoryIDs, tags, imagePath, primaryImageIndex)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Перевіримо, чи пост був створений
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM posts WHERE title = ?", title).Scan(&count)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 post, got %d", count)
	}
}

func TestCreatePost_TooManyCategories(t *testing.T) {
	db, teardown := SetupTestDB(t)
	defer teardown()

	// Додати 4 категорії
	db.Exec("INSERT INTO categories (id, name) VALUES (1,'One'), (2,'Two'), (3,'Three'), (4,'Four')")

	err := handlers.CreatePost(
		db,
		1,
		"Test title",
		"Test content",
		time.Now(),
		[]int{1, 2, 3, 4},
		[]string{},
		[]string{},
		1,
	)

	if err == nil || err.Error() != "can select up to 3 categories only" {
		t.Errorf("expected error about too many categories, got: %v", err)
	}
}

func TestCreatePost_NonExistentCategory(t *testing.T) {
	db, teardown := SetupTestDB(t)
	defer teardown()

	err := handlers.CreatePost(
		db,
		1,
		"Test title",
		"Test content",
		time.Now(),
		[]int{1, 2, 999},
		[]string{},
		[]string{},
		1,
	)
	if err == nil || err.Error() != "one or more categories do not exist" {
		t.Errorf("expected error about non-existent category, got: %v", err)
	}
}
