package db

import (
	"database/sql"
	"fmt"
	"forum/internal/security"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func SetupDatabase() error {

	err := godotenv.Load()
	if err != nil {
		return err
	}

	fmt.Println("DB_PATH =", os.Getenv("DB_PATH"))
	fmt.Println("DB_ENCRYPTION_KEY", os.Getenv("DB_ENCRYPTION_KEY"))
	// Get database path from environment variable or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join("app", "database", "forum.db")
	}

	// Create database directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory '%s': %v", dbDir, err)
	}

	// Open (or create) database file
	key := os.Getenv("DB_ENCRYPTION_KEY")
	if key == "" {
		return fmt.Errorf("encryption key not set in DB_ENCRYPTION_KEY")
	}

	dsn := fmt.Sprintf("%s?_pragma_key=%s&_pragma_cipher_page_size=4096", dbPath, key)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	}

	// Enable foreign key support
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %v", err)
	}

	// SQL schema for table creation
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
    parent_comment_id INTEGER,  -- <-- this is where we store the parent comment ID or NULL
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
    user_id INTEGER NOT NULL,          -- who the notification is intended for
    type TEXT NOT NULL CHECK (type IN ('like', 'dislike', 'comment', 'reply')),
    post_id INTEGER,
    comment_id INTEGER,                -- for the answer
    actor_id INTEGER,                  -- who did the action (answer)
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

	// Execute schema to create tables
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	// Check if admin already exists in the users table
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check admin existence: %v", err)
	}

	if count == 0 {
		// If there is no admin, create one
		adminUsername := "admin"
		adminEmail := "admin@example.com"
		adminPassword := "admin123"
		passwordHash, _ := security.HashPassword(adminPassword)

		_, err := db.Exec(`INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, 'admin')`,
			adminUsername, adminEmail, passwordHash)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}

		fmt.Println("Admin user created with username 'admin' and password 'admin123'")
	}

	fmt.Printf("Database successfully initialized at: %s\n", dbPath)
	return nil
}

// DropTable drops a table if it exists
func DropTable(tableName string) error {
	// Use consistent database path
	dbCon, err := sql.Open("sqlite3", "app/database/forum.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer dbCon.Close()

	// Enable foreign keys
	_, err = dbCon.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %v", err)
	}

	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err = dbCon.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %v", tableName, err)
	}

	fmt.Printf("Table '%s' dropped successfully\n", tableName)
	return nil
}

// RecreateDatabase drops and recreates all tables (use with caution!)
func RecreateDatabase() error {
	// List of tables in dependency order (reverse order for dropping)
	tables := []string{"sessions", "likes", "post_categories", "comments", "posts", "categories", "users"}

	// Drop all tables
	for _, table := range tables {
		err := DropTable(table)
		if err != nil {
			log.Printf("Warning: failed to drop table %s: %v", table, err)
		}
	}

	// Recreate all tables
	return SetupDatabase()
}

// EnableForeignKeys enables foreign key constraints on a database connection
func EnableForeignKeys(db *sql.DB) error {
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %v", err)
	}
	return nil
}

// CheckForeignKeys checks if foreign keys are enabled
func CheckForeignKeys(db *sql.DB) (bool, error) {
	var enabled bool
	err := db.QueryRow("PRAGMA foreign_keys").Scan(&enabled)
	if err != nil {
		return false, fmt.Errorf("failed to check foreign keys: %v", err)
	}
	return enabled, nil
}

// TestDatabaseIntegrity tests if the database setup is working correctly
func TestDatabaseIntegrity() error {
	db, err := sql.Open("sqlite3", "app/database/forum.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Enable foreign keys
	err = EnableForeignKeys(db)
	if err != nil {
		return err
	}

	// Check if foreign keys are enabled
	enabled, err := CheckForeignKeys(db)
	if err != nil {
		return err
	}

	fmt.Printf("Foreign keys enabled: %v\n", enabled)

	// Test foreign key constraints
	fmt.Println("Testing foreign key constraints...")

	// This should fail if constraints are working
	_, err = db.Exec("INSERT INTO posts (user_id, title, content) VALUES (999, 'Test', 'Test')")
	if err != nil {
		fmt.Println("✓ Foreign key constraint working - cannot insert post with non-existent user")
	} else {
		fmt.Println("✗ Warning: Foreign key constraints not working properly")
	}

	return nil
}
