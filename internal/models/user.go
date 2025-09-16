package models

type NewUser struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    string
	Banned       bool
	Role         string // guest, user, moderator, admin
}

type UserPageData struct {
	Posts       []PostView
	CurrentUser *User
	Categories  []Category
}

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    string
	AvatarPath   string
	CreatedPosts []Post
	LikedPosts   []LikedPosts
	DislikePosts []DislikePosts
	Likes        int
	Dislikes     int
	Banned       bool
	Role         string // guest, user, moderator, admin
}

type ProfilePageData struct {
	User             User
	CurrentUser      *User
	RequestSent      bool   // show successful submission message
	RequestError     string // show error if any
	Notifications    []Notification
	PostsWithComment []PostView
}

type LikedPosts struct {
	ID    string
	Title string
}
type DislikePosts struct {
	ID    string
	Title string
}
