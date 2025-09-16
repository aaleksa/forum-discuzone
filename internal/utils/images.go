package utils

import (
	"context"
	"database/sql"
	"fmt"
	"forum/internal/models"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
)

func ProcessPostImages(ctx context.Context, tx *sql.Tx, postID int64, imagePaths []string, primaryImageIndex int) error {
	// Добавляем детальное логирование
	log.Printf("ProcessPostImages called with postID=%d, imagePaths=%+v, primaryImageIndex=%d", postID, imagePaths, primaryImageIndex)

	// Validate inputs
	if postID <= 0 {
		log.Printf("Error: invalid post ID: %d", postID)
		return fmt.Errorf("invalid post ID")
	}

	if len(imagePaths) == 0 {
		log.Printf("No images to process for post %d", postID)
		return nil
	}

	// Validate primary image index
	if primaryImageIndex < 0 || primaryImageIndex >= len(imagePaths) {
		log.Printf("Error: primary image index out of range (got %d, max %d)", primaryImageIndex, len(imagePaths)-1)
		return fmt.Errorf("primary image index out of range (got %d, max %d)", primaryImageIndex, len(imagePaths)-1)
	}

	// Check that the post_images table exists
	var tableExists int
	err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='post_images'").Scan(&tableExists)
	if err != nil {
		log.Printf("Error checking if post_images table exists: %v", err)
		return fmt.Errorf("failed to check table existence: %w", err)
	}
	if tableExists == 0 {
		log.Printf("Error: post_images table does not exist")
		return fmt.Errorf("post_images table does not exist")
	}

	// First, we delete old entries for this post (if any)
	deleteQuery := "DELETE FROM post_images WHERE post_id = ?"
	deleteResult, err := tx.ExecContext(ctx, deleteQuery, postID)
	if err != nil {
		log.Printf("Error deleting old images for post %d: %v", postID, err)
		return fmt.Errorf("failed to delete old images: %w", err)
	}
	rowsDeleted, _ := deleteResult.RowsAffected()
	log.Printf("Deleted %d old image records for post %d", rowsDeleted, postID)

	// Prepare statement для вставки
	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO post_images 
        (post_id, image_path, is_primary, order_index) 
        VALUES (?, ?, ?, ?)`)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert each image
	for i, path := range imagePaths {
		if path == "" {
			log.Printf("Warning: empty image path at index %d", i)
			continue
		}

		isPrimary := i == primaryImageIndex
		log.Printf("Inserting image %d: path=%s, isPrimary=%t, orderIndex=%d", i, path, isPrimary, i)

		result, err := stmt.ExecContext(ctx,
			postID,
			path,
			isPrimary,
			i,
		)
		if err != nil {
			log.Printf("Error inserting image %d (%s): %v", i, path, err)
			return fmt.Errorf("failed to insert image %d (%s): %w", i, path, err)
		}

		// We check that the record was inserted
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Error getting rows affected for image %d: %v", i, err)
		} else {
			log.Printf("Successfully inserted image %d for post %d, rows affected: %d", i, postID, rowsAffected)
		}
	}

	// We check that the records were indeed inserted
	var insertedCount int
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM post_images WHERE post_id = ?", postID).Scan(&insertedCount)
	if err != nil {
		log.Printf("Error checking inserted images count: %v", err)
	} else {
		log.Printf("Total images in database for post %d: %d", postID, insertedCount)
	}

	return nil
}

// GetImagesByPostID fetches image paths associated with a specific post
func GetPostImages(db *sql.DB, postID int) ([]models.Image, error) {
	rows, err := db.Query(`
		SELECT image_path, is_primary, order_index, id
		FROM post_images
		WHERE post_id = ?
		ORDER BY order_index ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		err := rows.Scan(&img.Path, &img.IsPrimary, &img.Order, &img.ID) // додано &img.ID
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

// Добавьте эту функцию для проверки схемы базы данных
func CheckDatabaseSchema(db *sql.DB) {
	log.Println("=== Checking database schema ===")

	// Проверяем существование таблицы post_images
	var tableExists int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='post_images'").Scan(&tableExists)
	if err != nil {
		log.Printf("Error checking table existence: %v", err)
		return
	}

	if tableExists == 0 {
		log.Println("ERROR: post_images table does not exist!")
		log.Println("You need to create it with:")
		log.Println(`CREATE TABLE post_images (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            post_id INTEGER NOT NULL,
            image_path TEXT NOT NULL,
            is_primary BOOLEAN DEFAULT FALSE,
            order_index INTEGER DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
        );`)
		return
	}

	log.Println("post_images table exists")

	// Получаем структуру таблицы
	rows, err := db.Query("PRAGMA table_info(post_images)")
	if err != nil {
		log.Printf("Error getting table info: %v", err)
		return
	}
	defer rows.Close()

	log.Println("post_images table structure:")
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		log.Printf("  Column: %s, Type: %s, NotNull: %d, PK: %d", name, dataType, notNull, pk)
	}

	// Проверяем содержимое таблицы
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM post_images").Scan(&count)
	if err != nil {
		log.Printf("Error counting post_images: %v", err)
	} else {
		log.Printf("Total records in post_images: %d", count)
	}

	log.Println("=== End schema check ===")
}

// processUploadedImages processes uploaded files and returns slice of saved image paths and primary image index.
func ProcessUploadedImages(r *http.Request) ([]string, int, error) {
	var imagePaths []string
	primaryImageIndex := 0

	if idxStr := r.FormValue("primaryImageIndex"); idxStr != "" {
		if i, err := strconv.Atoi(idxStr); err == nil && i >= 0 {
			primaryImageIndex = i
			log.Printf("Primary image index set to: %d", primaryImageIndex)
		} else {
			log.Printf("Invalid primaryImageIndex value: %s, defaulting to 0", idxStr)
		}
	}

	possibleFieldNames := []string{"images[]", "images", "image[]", "image", "files[]", "files"}
	var foundFiles []*multipart.FileHeader

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, 0, fmt.Errorf("No files uploaded")
	}

	for _, fieldName := range possibleFieldNames {
		if files, exists := r.MultipartForm.File[fieldName]; exists && len(files) > 0 {
			foundFiles = files
			log.Printf("Found %d files in field '%s'", len(files), fieldName)
			break
		}
	}

	if len(foundFiles) == 0 {
		log.Printf("No files found in any expected field. Available fields: %+v", r.MultipartForm.File)
		return nil, primaryImageIndex, nil // No error, just no images
	}

	for idx, fileHeader := range foundFiles {
		log.Printf("Processing file %d/%d: %s (%d bytes)", idx+1, len(foundFiles), fileHeader.Filename, fileHeader.Size)

		if fileHeader.Size == 0 || fileHeader.Filename == "" {
			log.Printf("Skipping empty or unnamed file")
			continue
		}

		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Failed to open file %s: %v", fileHeader.Filename, err)
			continue
		}
		defer file.Close()

		imagePath, err := SaveUploadedFile(file, fileHeader)
		if err != nil {
			log.Printf("Failed to save file %s: %v", fileHeader.Filename, err)
			switch err.Error() {
			case "file is too large":
				return nil, 0, fmt.Errorf("Maximum allowed size is 20MB")
			case "unsupported file type":
				return nil, 0, fmt.Errorf("Only JPEG, PNG, and GIF are allowed")
			default:
				continue
			}
		}

		imagePaths = append(imagePaths, imagePath)
		log.Printf("Saved file to: %s", imagePath)
	}

	return imagePaths, primaryImageIndex, nil
}
