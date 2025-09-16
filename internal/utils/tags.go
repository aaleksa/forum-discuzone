package utils

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"strings"
)

func GetPostTags(db *sql.DB, postID int) ([]string, error) {
	query := `
        SELECT t.name 
        FROM tags t
        JOIN post_tags pt ON t.id = pt.tag_id
        WHERE pt.post_id = ?`

	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// UpdatePostTags updates the tags for a post
func UpdatePostTags(db *sql.DB, postID int, newTags []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// First, remove all existing tags for this post
	_, err = tx.Exec("DELETE FROM post_tags WHERE post_id = ?", postID)
	if err != nil {
		return fmt.Errorf("failed to remove old tags: %w", err)
	}

	// Then add the new tags
	if len(newTags) > 0 {
		// Prepare statements
		insertTagStmt, err := tx.Prepare(`
            INSERT OR IGNORE INTO tags (name) VALUES (?)`)
		if err != nil {
			return fmt.Errorf("prepare tag insert: %w", err)
		}
		defer insertTagStmt.Close()

		getTagStmt, err := tx.Prepare(`SELECT id FROM tags WHERE name = ?`)
		if err != nil {
			return fmt.Errorf("prepare tag select: %w", err)
		}
		defer getTagStmt.Close()

		linkTagStmt, err := tx.Prepare(`INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)`)
		if err != nil {
			return fmt.Errorf("prepare tag link: %w", err)
		}
		defer linkTagStmt.Close()

		// Process each tag
		for _, tagName := range newTags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			// Insert tag if new
			if _, err := insertTagStmt.Exec(tagName); err != nil {
				return fmt.Errorf("insert tag: %w", err)
			}

			// Get tag ID
			var tagID int64
			if err := getTagStmt.QueryRow(tagName).Scan(&tagID); err != nil {
				return fmt.Errorf("get tag ID: %w", err)
			}

			// Link tag to post
			if _, err := linkTagStmt.Exec(postID, tagID); err != nil {
				return fmt.Errorf("link tag: %w", err)
			}
		}
	}

	return tx.Commit()
}

func ProcessPostTags(ctx context.Context, tx *sql.Tx, postID int64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	insertTagStmt, err := tx.PrepareContext(ctx, `
        INSERT OR IGNORE INTO tags (name) VALUES (?)`)
	if err != nil {
		log.Printf("Tag insert statement error: %v", err)
		return fmt.Errorf("tag insert statement error: %w", err)
	}
	defer insertTagStmt.Close()

	getTagStmt, err := tx.PrepareContext(ctx, `
        SELECT id FROM tags WHERE name = ?`)
	if err != nil {
		log.Printf("Tag select statement error: %v", err)
		return fmt.Errorf("tag select statement error: %w", err)
	}
	defer getTagStmt.Close()

	linkTagStmt, err := tx.PrepareContext(ctx, `
        INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)`)
	if err != nil {
		log.Printf("Tag link statement error: %v", err)
		return fmt.Errorf("tag link statement error: %w", err)
	}
	defer linkTagStmt.Close()

	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}

		log.Printf("Processing tag: '%s'", tagName)

		if _, err := insertTagStmt.ExecContext(ctx, tagName); err != nil {
			log.Printf("Tag '%s' insert error: %v", tagName, err)
			return fmt.Errorf("tag '%s' insert error: %w", tagName, err)
		}

		var tagID int64
		if err := getTagStmt.QueryRowContext(ctx, tagName).Scan(&tagID); err != nil {
			log.Printf("Tag '%s' ID retrieval error: %v", tagName, err)
			return fmt.Errorf("tag '%s' ID retrieval error: %w", tagName, err)
		}

		log.Printf("Linking tag %d ('%s') to post %d", tagID, tagName, postID)

		if _, err := linkTagStmt.ExecContext(ctx, postID, tagID); err != nil {
			log.Printf("Tag link error (post %d, tag %d): %v", postID, tagID, err)
			return fmt.Errorf("tag link error: %w", err)
		}
	}

	log.Printf("Added %d tags to post %d", len(tags), postID)
	return nil
}
