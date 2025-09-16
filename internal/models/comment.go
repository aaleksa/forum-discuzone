package models

type Comment struct {
	ID              int
	PostID          int
	UserID          int
	UserName        string
	Content         string
	Likes           int
	Dislikes        int
	CreatedAt       string
	ParentCommentID int
	Username        string
	PostTitle       string
}
