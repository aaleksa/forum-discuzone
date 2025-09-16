package models

type ModerationRequest struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Username    string `json:"username"` // Add this field
	Status      string `json:"status"`
	RequestedAt string `json:"requested_at"`
	ReviewedAt  string `json:"reviewed_at,omitempty"`
	ReviewedBy  int    `json:"reviewed_by,omitempty"`
}
