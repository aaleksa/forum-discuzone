package models

type FilterPageData struct {
	Posts              []PostView
	CurrentUser        *User
	Categories         []Category
	CurrentFilter      string
	SelectedCategories []int
}
