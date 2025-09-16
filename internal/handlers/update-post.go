package handlers

import (
	"database/sql"
	"forum/internal"
	"forum/internal/models"
	"forum/internal/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func EditPostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleEditGet(w, r, db)
		case http.MethodPost:
			handleEditPost(w, r, db)
		default:
			errors.RenderError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "The HTTP method is not supported.")
		}
	}
}

func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func handleEditGet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("DEBUG: Starting handleEditGet")

	// Parse post ID from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		log.Printf("ERROR: Invalid URL format: %s", r.URL.Path)
		errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid URL formatprovided.")
		return
	}

	postID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Printf("ERROR: Invalid post ID: %v (from path: %s)", err, r.URL.Path)
		errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid post ID.")
		return
	}
	log.Printf("DEBUG: Parsed post ID: %d", postID)

	// Get categories
	categories, err := utils.GetCategories(w, db)
	if err != nil {
		log.Printf("ERROR: Failed to get categories: %v", err)
		errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Internal server error.")
		return
	}
	log.Println("DEBUG: Fetched categories")

	// Check user session
	currentUser, err := utils.GetUserFromSession(w, r, db)
	if err != nil {
		log.Printf("ERROR: Session check error: %v", err)
		errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Session check error.")
		return
	}
	log.Printf("DEBUG: Current user from session: %s", currentUser.Username)

	// Get the post
	post, err := utils.GetPostByID(db, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ERROR: Post not found (ID: %d)", postID)
			errors.RenderError(w, http.StatusNotFound, "Not Found", "Post not found.")
		} else {
			log.Printf("ERROR: Database error while fetching post: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Something went wrong post.")
		}
		return
	}
	log.Printf("DEBUG: Fetched post: %+v", post)

	// Get the post's current categories
	postCategories, err := utils.GetPostCategories(db, postID)
	if err != nil {
		log.Printf("ERROR: Failed to get post categories: %v", err)
		errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Something went wrong categories.")
		return
	}
	log.Printf("DEBUG: Fetched post categories: %+v", postCategories)

	// Create selected categories map using simple map[int]bool
	selectedCategories := make(map[int]bool)
	for _, categoryID := range postCategories {
		selectedCategories[categoryID] = true
	}
	log.Printf("DEBUG: Selected categories map: %+v", selectedCategories)

	// Get the post's current tags
	postTags, err := utils.GetPostTags(db, postID)
	if err != nil {
		log.Printf("ERROR: Failed to get post tags: %v", err)
		errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Something went wrong tags.")
		return
	}
	log.Printf("DEBUG: Fetched post tags: %+v", postTags)

	// Convert tags slice to comma-separated string for the form
	tagsString := strings.Join(postTags, ", ")

	// Parsing templates
	tmpl := template.New("layout.html").Funcs(template.FuncMap{
		"containsString": containsString,
		"join":           strings.Join,
	})

	tmpl, err = tmpl.ParseFiles(
		"templates/layout.html",
		"templates/edit_post.html",
		"templates/header.html",
		"templates/nav.html",
		"templates/images-post.html",
		"templates/notifications.html",
	)

	if err != nil {
		log.Printf("ERROR: Template parsing error: %v", err)
		errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Something went wrong Template.")
		return
	}
	log.Println("DEBUG: Templates parsed successfully")

	// Form data for the template
	data := models.UpdatePostPageData{
		Post:               post,
		SelectedCategories: selectedCategories,
		Categories:         categories,
		CurrentUser:        currentUser,
		Tags:               tagsString,
		CSRFToken:          "example-token",
	}

	// Execute the template
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ERROR: Template execution error: %v", err)
		errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Something went wrong Template.")
	} else {
		log.Println("DEBUG: Template executed successfully")
	}
}

func handleEditPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid form data.")
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	idStr := r.FormValue("id")
	tags := r.FormValue("tags")

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid post ID.")
		return
	}

	// Get current user from session
	currentUser, err := utils.GetUserFromSession(w, r, db)
	if err != nil || currentUser == nil {
		log.Printf("Unauthorized delete attempt: %v", err)
		errors.RenderError(w, http.StatusUnauthorized, "Unauthorized", "Please log in.")
		return
	}

	post, err := utils.GetPostByID(db, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RenderError(w, http.StatusNotFound, "Not Found", "Post not found.")
		} else {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to get post.")
		}
		return
	}

	// Check permissions (admin, moderator, or post author)
	if !utils.HasPermission(currentUser, post.UserID, "edit") {
		errors.RenderError(w, http.StatusForbidden, "Forbidden", "You don't have permission to edit this post.")
		return
	}

	// === Categories ===
	categories := r.Form["categories[]"]
	if len(categories) == 0 {
		categories = r.Form["categories"]
	}

	// === Processing multiple images ===
	var newImagePaths []string
	files := r.MultipartForm.File["images[]"]
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			log.Printf("Failed to open uploaded file: %v", err)
			continue
		}
		defer file.Close()

		path, err := utils.SaveUploadedFile(file, fh)
		if err != nil {
			log.Printf("Failed to save file: %v", err)
			continue
		}
		newImagePaths = append(newImagePaths, path)
	}

	// === Processing IDs of images to be deleted ===
	var removeImageIDs []int
	removeIDs := r.Form["remove_images[]"]
	for _, idStr := range removeIDs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Invalid image ID to remove: %s", idStr)
			continue
		}
		removeImageIDs = append(removeImageIDs, id)
	}

	// === Post update ===
	err = utils.UpdatePostFull(
		r.Context(),
		db,
		postID,
		title,
		content,
		categories,
		tags,
		newImagePaths,
		removeImageIDs,
	)
	if err != nil {
		log.Printf("UpdatePostFull error: %v", err)
		errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to update post.")
		return
	}

	http.Redirect(w, r, "/user_page", http.StatusFound)
}
