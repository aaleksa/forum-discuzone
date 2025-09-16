package models

type SearchPageData struct {
	Query       string
	Results     []PostView
	Message     string
	CurrentUser *User
}
