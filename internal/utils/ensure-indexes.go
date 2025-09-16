package utils

import (
	"database/sql"
	"log"
)

func EnsureIndexes(db *sql.DB) {
	indexes := []string{
		// posts
		`CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);`,

		// comments
		`CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);`,

		// post_categories
		`CREATE INDEX IF NOT EXISTS idx_post_categories_post_id ON post_categories(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_post_categories_category_id ON post_categories(category_id);`,

		// likes (окремі унікальні індекси замість проблемного трійкового UNIQUE)
		`CREATE UNIQUE INDEX IF NOT EXISTS unique_like_post ON likes(user_id, post_id) WHERE post_id IS NOT NULL;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS unique_like_comment ON likes(user_id, comment_id) WHERE comment_id IS NOT NULL;`,
		`CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);`,

		// sessions
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);`,

		// tags
		`CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);`,

		// post_tags
		`CREATE INDEX IF NOT EXISTS idx_post_tags_post_id ON post_tags(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_post_tags_tag_id ON post_tags(tag_id);`,

		// moderator_requests
		`CREATE INDEX IF NOT EXISTS idx_moderator_requests_user_id ON moderator_requests(user_id);`,
	}

	for _, query := range indexes {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("Index creation failed: %v", err)
		}
	}
}
