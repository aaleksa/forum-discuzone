package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"forum/internal"
	"forum/internal/models"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

var (
	posts []models.Post
	mu    sync.Mutex
)

func ServeFormCreatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categories, err := utils.GetCategories(w, db)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to get categories")
			return
		}

		CurrentUser, err := utils.GetUserFromSession(w, r, db)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Session check error")
			return
		}

		data := models.CreatePostPageData{
			Categories:  categories,
			CurrentUser: CurrentUser,
			CSRFToken:   "example-token",
		}

		tmpl, err := template.ParseFiles(
			"templates/layout.html",
			"templates/header.html",
			"templates/nav.html",
			"templates/create_post.html",
			"templates/form_group_post.html",
			"templates/images-post.html",
			"templates/notifications.html",
		)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template error: "+err.Error())
			return
		}

		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Render error: "+err.Error())
		}
	}
}

func HandlerCreatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(50 << 20) // 50MB max
		if err != nil {
			log.Printf("Error parsing multipart form: %v", err)
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Error parsing form.")
			return
		}

		// Get userID from context
		userID, ok := utils.MustGetUserID(w, r)
		if !ok {
			return // Return nil for both user and error
		}

		title := r.FormValue("title")
		content := r.FormValue("content")
		createdAt := time.Now()
		tagsInput := r.FormValue("tags")

		// Process categories
		categoryIDs, err := parseCategoryIDs(r)
		if err != nil {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", err.Error())
			return
		}

		// Process tags
		tags := parseTags(tagsInput)

		// Process images
		imagePaths, primaryImageIndex, err := utils.ProcessUploadedImages(r)
		if err != nil {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", err.Error())
			return
		}

		// Validate primary image index
		if len(imagePaths) == 0 && primaryImageIndex != 0 {
			log.Printf("Warning: primaryImageIndex set but no images uploaded; resetting to 0")
			primaryImageIndex = 0
		} else if primaryImageIndex < 0 || primaryImageIndex >= len(imagePaths) {
			log.Printf("Warning: invalid primaryImageIndex %d; resetting to 0", primaryImageIndex)
			primaryImageIndex = 0
		}

		err = CreatePost(db, userID, title, content, createdAt, categoryIDs, tags, imagePaths, primaryImageIndex)
		if err != nil {
			log.Printf("Error creating post: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Error adding post to database.")
			return
		}

		log.Printf("Post created successfully with %d images", len(imagePaths))
		http.Redirect(w, r, "/user_page", http.StatusSeeOther)
	}
}

// parseCategoryIDs extracts category IDs from request form, ensures they are unique and valid.
func parseCategoryIDs(r *http.Request) ([]int, error) {
	categories := r.Form["categories[]"]
	if len(categories) == 0 {
		categories = r.Form["categories"]
	}
	if len(categories) == 0 {
		categories = r.PostForm["categories[]"]
	}
	if len(categories) == 0 {
		categories = r.PostForm["categories"]
	}

	var categoryIDs []int
	seenIDs := make(map[int]bool)
	for _, category := range categories {
		category = strings.TrimSpace(category)
		if category == "" {
			continue
		}

		id, err := strconv.Atoi(category)
		if err != nil {
			return nil, fmt.Errorf("Invalid category ID: %s", category)
		}

		if seenIDs[id] {
			return nil, fmt.Errorf("Duplicate categories found.")
		}
		seenIDs[id] = true
		categoryIDs = append(categoryIDs, id)
	}

	return categoryIDs, nil
}

// parseTags splits and trims tag input string into a slice of tags.
func parseTags(tagsInput string) []string {
	if tagsInput == "" {
		return nil
	}

	tags := strings.Split(tagsInput, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return tags
}

func CreatePost(db *sql.DB, userID int, title, content string, created_at time.Time, categoryIDs []int, tags []string, imagePaths []string, primaryImageIndex int) error {
	// Start of transaction
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	// Ensure transaction is rolled back on any error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Log the values being inserted
	log.Printf("Inserting post: userID=%d, title=%s, content=%s, created_at=%v", userID, title, content, created_at, imagePaths)
	log.Printf("Categories to insert: %+v", categoryIDs)

	// Adding a post
	query := `INSERT INTO posts (user_id, title, content, created_at) VALUES (?, ?, ?, ?)`
	result, err := tx.Exec(query, userID, title, content, created_at)
	if err != nil {
		log.Printf("Error inserting post: %v", err)
		return err
	}

	// Getting the ID of a new post
	postID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		return err
	}

	log.Printf("New post ID: %d", postID)
	// Check maximum number of categories (3)
	if len(categoryIDs) > 3 {
		return fmt.Errorf("can select up to 3 categories only")
	}

	if len(categoryIDs) > 0 {
		valid, err := validateCategoriesExist(tx, categoryIDs)
		if err != nil {
			log.Printf("Error validating categories: %v", err)
			return err
		}
		if !valid {
			return fmt.Errorf("one or more categories do not exist")
		}

		// Delete old categories (if any)
		if _, err := tx.Exec("DELETE FROM post_categories WHERE post_id = ?", postID); err != nil {
			log.Printf("Error deleting old categories: %v", err)
			return err
		}

		stmt, err := tx.Prepare("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)")
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			return err
		}
		defer stmt.Close()

		for _, categoryID := range categoryIDs {
			if _, err := stmt.Exec(postID, categoryID); err != nil {
				log.Printf("Error adding category %d to post %d: %v", categoryID, postID, err)
				return err
			}
			log.Printf("Post %d linked to category %d", postID, categoryID)
		}
	} else {
		log.Printf("No categories selected for post %d", postID)
	}

	// Calling a function to process tags
	err = utils.ProcessPostTags(ctx, tx, postID, tags)
	if err != nil {
		log.Printf("Error processing tags for post %d: %v", postID, err)
		return err
	}

	// Calling a function to process images
	err = utils.ProcessPostImages(ctx, tx, postID, imagePaths, primaryImageIndex)
	if err != nil {
		log.Printf("Error processing images for post %d: %v", postID, err)
		return err
	}

	// Completion of the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Println("Post and categories added successfully")
	return nil

}

// Auxiliary function for checking the existence of categories
func validateCategoriesExist(tx *sql.Tx, ids []int) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}

	query := "SELECT COUNT(*) FROM categories WHERE id IN (?" + strings.Repeat(",?", len(ids)-1) + ")"
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	var count int
	err := tx.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == len(ids), nil
}
