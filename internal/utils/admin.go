package utils

import (
	"database/sql"
	"fmt"
	"forum/internal/models"
	"log"
)

// Helper function for receiving moderation requests
func GetModerationRequests(db *sql.DB) ([]models.ModerationRequest, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
        SELECT mr.id, mr.user_id, u.username, mr.status, 
               strftime('%Y-%m-%d %H:%M:%S', mr.requested_at) as requested_at,
               CASE WHEN mr.reviewed_at IS NULL THEN '' ELSE strftime('%Y-%m-%d %H:%M:%S', mr.reviewed_at) END as reviewed_at,
               COALESCE(mr.reviewed_by, 0) as reviewed_by
        FROM moderator_requests mr
        JOIN users u ON mr.user_id = u.id
        WHERE mr.status = 'pending'
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	var requests []models.ModerationRequest
	for rows.Next() {
		var req models.ModerationRequest
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.Username, // New field
			&req.Status,
			&req.RequestedAt,
			&req.ReviewedAt,
			&req.ReviewedBy,
		)
		if err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return requests, nil
}

func IsAdmin(user *models.User) bool {
	return user != nil && user.Role == "admin"
}
