package test

import (
	"database/sql"
	"os"
	"testing"
)

// --- Тестова база для всіх тестів ---
func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}

	// Створення всіх таблиць
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		role TEXT DEFAULT 'user',
		avatar_url TEXT DEFAULT NULL,
		banned BOOLEAN DEFAULT FALSE,
		provider TEXT DEFAULT '',
		provider_id TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME,
		image_path TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS post_images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		image_path TEXT NOT NULL,
		is_primary BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		order_index INTEGER DEFAULT 0,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		parent_comment_id INTEGER,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (parent_comment_id) REFERENCES comments(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS post_categories (
		post_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		PRIMARY KEY (post_id, category_id),
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER,
		comment_id INTEGER,
		reaction TEXT NOT NULL CHECK (reaction IN ('Like', 'Dislike')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		CHECK (post_id IS NOT NULL OR comment_id IS NOT NULL),
		UNIQUE(user_id, post_id, comment_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS post_tags (
		post_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (post_id, tag_id),
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS moderator_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		status TEXT CHECK(status IN ('pending', 'approved', 'rejected')) DEFAULT 'pending',
		requested_at TEXT NOT NULL,
		reviewed_at TEXT,
		reviewed_by INTEGER,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (reviewed_by) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		type TEXT NOT NULL CHECK (type IN ('like', 'dislike', 'comment', 'reply')),
		post_id INTEGER,
		comment_id INTEGER,
		actor_id INTEGER,
		is_read BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (comment_id) REFERENCES comments(id),
		FOREIGN KEY (actor_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

	CREATE TABLE IF NOT EXISTS password_resets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Додавання базових тестових даних
	_, _ = db.Exec(`
-- Користувачі
INSERT INTO users (username, email, password, role) VALUES 
('alice', 'alice@example.com', 'password', 'user'),
('bob', 'bob@example.com', 'password', 'moderator');

-- Категорії
INSERT INTO categories (name) VALUES 
('Technology'), 
('Science');

-- Теги
INSERT INTO tags (name) VALUES 
('Go'), 
('Programming');

-- Пости
INSERT INTO posts (user_id, title, content, created_at) VALUES 
(1, 'Hello World', 'This is a test post', CURRENT_TIMESTAMP),
(2, 'Second Post', 'Another test post', CURRENT_TIMESTAMP);

-- Зображення постів
INSERT INTO post_images (post_id, image_path, is_primary, order_index) VALUES
(1, '/images/post1.png', 1, 0),
(2, '/images/post2.png', 1, 0);

-- Пост-категорії
INSERT INTO post_categories (post_id, category_id) VALUES
(1, 1),
(2, 2);

-- Пост-теги
INSERT INTO post_tags (post_id, tag_id) VALUES
(1, 1),
(2, 2);

-- Коментарі
INSERT INTO comments (post_id, user_id, content, created_at) VALUES
(1, 2, 'Nice post!', CURRENT_TIMESTAMP),
(1, 1, 'Thank you!', CURRENT_TIMESTAMP);

-- Лайки
INSERT INTO likes (user_id, post_id, reaction) VALUES
(1, 1, 'Like'),
(2, 1, 'Like');

-- Сесії
INSERT INTO sessions (id, user_id, created_at, expires_at) VALUES
('session1', 1, CURRENT_TIMESTAMP, DATETIME('now', '+1 hour')),
('session2', 2, CURRENT_TIMESTAMP, DATETIME('now', '+1 hour'));

-- Запити модераторів
INSERT INTO moderator_requests (user_id, status, requested_at) VALUES
(1, 'pending', CURRENT_TIMESTAMP),
(2, 'approved', CURRENT_TIMESTAMP);

-- Повідомлення
INSERT INTO notifications (user_id, type, post_id, comment_id, actor_id) VALUES
(1, 'like', 1, NULL, 2),
(2, 'comment', 1, 1, 1);

-- Скидання пароля
INSERT INTO password_resets (user_id, token, expires_at) VALUES
(1, 'reset-token-123', DATETIME('now', '+1 day')),
(2, 'reset-token-456', DATETIME('now', '+1 day'));
`)

	teardown := func() {
		db.Close()
		os.Remove("file::memory:")
	}

	return db, teardown
}
