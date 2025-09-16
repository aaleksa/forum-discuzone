package models

type UserActivityResults struct {
	Posts    []PostView
	Comments []Comment
	Likes    []PostView
}

type ActivitySearchPageData struct {
	Query            string
	FilterType       string
	Results          UserActivityResults
	Message          string
	User             User
	CurrentUser      *User
	Notifications    []Notification
	PostsWithComment []PostView
}
