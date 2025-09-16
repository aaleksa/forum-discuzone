package test

import (
	"fmt"
	"forum/internal/utils"
	"testing"
)

func TestSearchPosts(t *testing.T) {
	db, teardown := SetupTestDB(t)
	defer teardown()

	posts, err := utils.SearchPosts(db, "Hello")
	if err != nil {
		t.Fatalf("SearchPosts returned error: %v", err)
	}

	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	if posts[0].Title != "Hello World" {
		t.Errorf("expected title 'Hello World', got '%s'", posts[0].Title)
	}
}

func TestSearchUserActivity(t *testing.T) {
	db, teardown := SetupTestDB(t)
	defer teardown()

	// Clearing tables in the correct order via FK
	tables := []string{"likes", "comments", "posts", "users"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Fatalf("Failed to clear table %s: %v", table, err)
		}
	}

	// Inserting test data
	_, err := db.Exec(`
	INSERT INTO users (id, username, email, password) VALUES
		(1, 'Alice', 'alice@example.com', 'pass1'),
		(2, 'Bob', 'bob@example.com', 'pass2');

	INSERT INTO posts (id, user_id, title, content) VALUES
		(1, 1, 'Hello World', 'This is my first post');

	INSERT INTO comments (id, post_id, user_id, content) VALUES
        (1, 1, 1, 'Nice Hello post!');

    INSERT INTO likes (id, post_id, user_id, reaction) VALUES
        (1, 1, 1, 'Like');

	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Calling the tested function
	results, err := utils.SearchUserActivity(db, 1, "Hello", "")
	if err != nil {
		t.Fatalf("SearchUserActivity returned error: %v", err)
	}

	// Checking the results
	if len(results.Posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(results.Posts))
	}
	if len(results.Comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(results.Comments))
	}
	if len(results.Likes) != 1 {
		t.Errorf("expected 1 like, got %d", len(results.Likes))
	}

	if len(results.Posts) > 0 && results.Posts[0].Title != "Hello World" {
		t.Errorf("expected post title 'Hello World', got '%s'", results.Posts[0].Title)
	}
	if len(results.Comments) > 0 && results.Comments[0].Content != "Nice Hello post!" {
		t.Errorf("expected comment content 'Nice Hello post!', got '%s'", results.Comments[0].Content)
	}
	if len(results.Likes) > 0 && results.Likes[0].Title != "Hello World" {
		t.Errorf("expected like to be on post 'Hello World', got '%s'", results.Likes[0].Title)
	}
}
