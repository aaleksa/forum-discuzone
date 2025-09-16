package utils

import (
	"bufio"
	"database/sql"
	"fmt"
	"forum/internal/models"
	"log"
	"net/http"
	"os"
	"strings"
)

func GetCategories(w http.ResponseWriter, db *sql.DB) ([]models.Category, error) {

	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// Function for adding new categories to the database
func UpdateCategories(db *sql.DB, categories []string) error {
	for _, category := range categories {
		// Checking if a category already exists
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE name = ?)", category).Scan(&exists)
		if err != nil {
			return err
		}

		// If the category does not exist, add it
		if !exists {
			_, err := db.Exec("INSERT INTO categories (name) VALUES (?)", category)
			if err != nil {
				return err
			}
			fmt.Printf("New category added: %s\n", category)
		}
	}

	return nil
}

// Function to read categories from a file
func ReadCategoriesFromFile(filename string) ([]string, error) {
	var categories []string

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			categories = append(categories, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// GetPostCategories retrieves the category IDs associated with a specific post
func GetPostCategories(db *sql.DB, postID int) ([]int, error) {
	log.Printf("DEBUG: Getting categories for post ID: %d", postID)

	// SQL query to get all category IDs for a specific post
	query := `
		SELECT category_id 
		FROM post_categories 
		WHERE post_id = ?
		ORDER BY category_id`

	// Execute the query
	rows, err := db.Query(query, postID)
	if err != nil {
		log.Printf("ERROR: Database query error in GetPostCategories: %v", err)
		return nil, fmt.Errorf("failed to query post categories: %w", err)
	}
	defer rows.Close()

	// Slice to store category IDs
	var categoryIDs []int

	// Iterate through the results
	for rows.Next() {
		var categoryID int
		if err := rows.Scan(&categoryID); err != nil {
			log.Printf("ERROR: Row scan error in GetPostCategories: %v", err)
			return nil, fmt.Errorf("failed to scan category ID: %w", err)
		}
		categoryIDs = append(categoryIDs, categoryID)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Rows iteration error in GetPostCategories: %v", err)
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	log.Printf("DEBUG: Found %d categories for post %d: %+v", len(categoryIDs), postID, categoryIDs)
	return categoryIDs, nil
}

// Alternative version that also returns category details (ID and Name)
func GetPostCategoriesWithDetails(db *sql.DB, postID int) ([]models.Category, error) {
	log.Printf("DEBUG: Getting category details for post ID: %d", postID)

	// SQL query to get category details for a specific post
	query := `
		SELECT c.id, c.name 
		FROM categories c
		INNER JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
		ORDER BY c.name`

	// Execute the query
	rows, err := db.Query(query, postID)
	if err != nil {
		log.Printf("ERROR: Database query error in GetPostCategoriesWithDetails: %v", err)
		return nil, fmt.Errorf("failed to query post categories with details: %w", err)
	}
	defer rows.Close()

	// Slice to store category details
	var categories []models.Category

	// Iterate through the results
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			log.Printf("ERROR: Row scan error in GetPostCategoriesWithDetails: %v", err)
			return nil, fmt.Errorf("failed to scan category details: %w", err)
		}
		categories = append(categories, category)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Rows iteration error in GetPostCategoriesWithDetails: %v", err)
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	log.Printf("DEBUG: Found %d category details for post %d", len(categories), postID)
	return categories, nil
}

// Helper function to check if a post has a specific category
func PostHasCategory(db *sql.DB, postID, categoryID int) (bool, error) {
	log.Printf("DEBUG: Checking if post %d has category %d", postID, categoryID)

	query := `
		SELECT COUNT(*) 
		FROM post_categories 
		WHERE post_id = ? AND category_id = ?`

	var count int
	err := db.QueryRow(query, postID, categoryID).Scan(&count)
	if err != nil {
		log.Printf("ERROR: Database query error in PostHasCategory: %v", err)
		return false, fmt.Errorf("failed to check post category: %w", err)
	}

	result := count > 0
	log.Printf("DEBUG: Post %d has category %d: %t", postID, categoryID, result)
	return result, nil
}
