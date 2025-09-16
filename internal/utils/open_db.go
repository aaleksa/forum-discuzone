package utils

import (
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"os"
)

// OpenDatabase opens a connection to the SQLite database and returns the connection.
func OpenDatabase() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	path := os.Getenv("DB_PATH")
	key := os.Getenv("DB_ENCRYPTION_KEY")

	dsn := path + "?_pragma_key=" + key + "&_pragma_cipher_page_size=4096"

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec("PRAGMA cipher_version;")
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
