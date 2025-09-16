package handlers

import (
	"database/sql"
	"fmt"
	"forum/internal"
	"forum/internal/models"
	"forum/internal/utils"
	"html/template"
	// "log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HandlePostsFilter(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Session check error.")
			return
		}

		categoryFilter := r.URL.Query()["categories[]"]
		likedOnly := r.URL.Query().Get("liked") == "true"
		myPosts := r.URL.Query().Get("mine") == "true"

		var joins []string
		var whereClauses []string
		var args []interface{}
		havingClause := ""

		if len(categoryFilter) > 0 {
			joins = append(joins, "JOIN post_categories pc ON p.id = pc.post_id")

			var catIDs []interface{}
			for _, catStr := range categoryFilter {
				catID, err := strconv.Atoi(catStr)
				if err != nil {
					errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid category ID.")
					return
				}
				catIDs = append(catIDs, catID)
			}

			placeholders := strings.Repeat("?,", len(catIDs))
			placeholders = strings.TrimSuffix(placeholders, ",")

			whereClauses = append(whereClauses, fmt.Sprintf("pc.category_id IN (%s)", placeholders))
			args = append(args, catIDs...)

			havingClause = fmt.Sprintf("HAVING COUNT(DISTINCT pc.category_id) = %d", len(catIDs))
		}

		if likedOnly {
			joins = append(joins, "JOIN likes l2 ON p.id = l2.post_id")
			whereClauses = append(whereClauses, "l2.user_id = ?")
			args = append(args, user.ID)
		}

		if myPosts {
			whereClauses = append(whereClauses, "p.user_id = ?")
			args = append(args, user.ID)
		}

		sqlQuery := `
        SELECT p.id, p.title, p.content, p.user_id, u.username as user_name,
               COUNT(DISTINCT l.id) as likes,
               COUNT(DISTINCT c.id) as comment_count,
               p.created_at
        FROM posts p
    `

		if len(joins) > 0 {
			sqlQuery += " " + strings.Join(joins, " ")
		}

		sqlQuery += `
        LEFT JOIN users u ON p.user_id = u.id
        LEFT JOIN likes l ON p.id = l.post_id
        LEFT JOIN comments c ON p.id = c.post_id
    `

		if len(whereClauses) > 0 {
			sqlQuery += " WHERE " + strings.Join(whereClauses, " AND ")
		}

		sqlQuery += " GROUP BY p.id "

		if havingClause != "" {
			sqlQuery += " " + havingClause
		}

		sqlQuery += " ORDER BY p.created_at DESC"

		fmt.Println("SQL Query:", sqlQuery)
		fmt.Println("Args:", args)

		rows, err := db.Query(sqlQuery, args...)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "DB query error: %v")
			return
		}
		defer rows.Close()

		var posts []models.PostView

		for rows.Next() {
			var p models.PostView
			var rawCreatedAt time.Time
			err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.UserID, &p.UserName, &p.Likes, &p.CommentsCount, &rawCreatedAt)
			if err != nil {
				errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Row scan error: %v")
				return
			}
			// Get tags for the post
			p.Tags, err = utils.GetPostTags(db, p.ID)
			if err != nil {
				errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", fmt.Sprintf("Tag fetch error: %v", err))
				return
			}
			// Format the date
			p.CreatedAt = utils.FormatDate(rawCreatedAt)
			// Receiving images
			p.ImagePaths, err = utils.GetPostImages(db, p.ID)
			if err != nil {
				errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", fmt.Sprintf("Images fetch error: %v", err))

			}

			posts = append(posts, p)
		}

		categories, err := utils.GetCategories(w, db)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to load categories: %v")
			return
		}
		currentFilter := ""
		if myPosts {
			currentFilter = "my"
		} else if likedOnly {
			currentFilter = "liked"
		}

		// Convert category IDs from string to int
		var selectedCategories []int
		for _, catStr := range categoryFilter {
			if catID, err := strconv.Atoi(catStr); err == nil {
				selectedCategories = append(selectedCategories, catID)
			}
		}

		data := models.FilterPageData{
			Posts:              posts,
			CurrentUser:        user,
			Categories:         categories,
			CurrentFilter:      currentFilter,
			SelectedCategories: selectedCategories,
		}

		tmpl, err := template.ParseFiles(
			"templates/layout.html",
			"templates/filters.html",
			"templates/filters_page.html",
			"templates/header.html",
			"templates/post_list_item.html",
			"templates/post_list_filter.html",
			"templates/nav.html",
			"templates/notifications.html",
		)
		if err != nil {
			fmt.Printf("Template parsing error: %v\n", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template error.")
			return
		}

		err = tmpl.ExecuteTemplate(w, "layout", data)
		if err != nil {
			fmt.Printf("Template execution error: %v\n", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template rendering error.")
			return
		}
	}
}
