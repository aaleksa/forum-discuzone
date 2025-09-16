package models

import (
	"database/sql"
)

// Struct for Post
type Post struct {
	ID        int
	UserID    int
	Title     string
	Likes     int
	Dislikes  int
	Content   string
	CreatedAt string
	Image     string
}

type CreatePostPageData struct {
	Categories  []Category
	CurrentUser *User
	CSRFToken   string
}
type UpdatePostPageData struct {
	Post               PostView
	Categories         []Category
	CurrentUser        *User
	CSRFToken          string
	Tags               string
	SelectedCategories map[int]bool
}
type Category struct {
	ID   int
	Name string
}

type PostPageData struct {
	Post             PostView
	Comments         []CommentWithReaction
	CurrentUser      *User
	UserReaction     string
	CommentReactions map[int]string
	Tags             []string
	CanModifyPost    bool
	CanModifyComment map[int]bool
	Replies          map[int][]Comment
}

type CommentWithReaction struct {
	Comment      // embedding your Comment
	UserReaction string
}

// Struct for Post view
type PostView struct {
	ID            int
	UserID        int
	CurrentUser   *User
	UserName      string
	Title         string
	Content       string
	Likes         int
	Dislikes      int
	IsLiked       bool
	CommentsCount int
	Comments      []Comment
	CreatedAt     string
	UpdatedAt     string
	IsEdited      bool
	Categories    []string
	Category      string
	Image         sql.NullString
	ImagePaths    []Image
	Tags          []string
}

type Image struct {
	ID        int
	Path      string
	IsPrimary bool
	Order     int
}
