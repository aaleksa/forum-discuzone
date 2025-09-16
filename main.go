package main

import (
	"database/sql"
	"forum/database"
	"forum/internal"
	"forum/internal/handlers"
	"forum/internal/middleware"
	"forum/internal/models"
	"forum/internal/utils"
	"github.com/joho/godotenv"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type App struct {
	DB *sql.DB
}

var databaseInitialized bool = false // Global variable to check if initialization has occurred

// NewApp opens the database, creates indexes, and returns App
func NewApp() (*App, error) {
	db, err := utils.OpenDatabase()
	if err != nil {
		return nil, err
	}

	utils.EnsureIndexes(db)

	return &App{DB: db}, nil
}

func init() {
	// Register correct MIME types
	mime.AddExtensionType(".js", "text/javascript")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".json", "application/json")
}

// Custom file server that ensures correct MIME types
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	// Remove the /static/ prefix
	path := strings.TrimPrefix(r.URL.Path, "/static/")

	// Security: prevent directory traversal
	if strings.Contains(path, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Construct full file path
	fullPath := filepath.Join("static", path)

	// Set correct content type based on file extension
	ext := filepath.Ext(fullPath)
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// Handler function to process requests
func (app *App) handler(w http.ResponseWriter, r *http.Request) {
	// Open the database connection
	db, err := utils.OpenDatabase()
	if err != nil {
		log.Printf("Error opening DB: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get current user from session
	currentUser, _ := utils.GetUserFromSession(w, r, app.DB) // Ignore error if not logged in

	// Reading categories from a file
	categoriesList, err := utils.ReadCategoriesFromFile("categories.txt")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	// Updating the database
	err = utils.UpdateCategories(db, categoriesList)
	if err != nil {
		log.Fatal("Error while updating database:", err)
	}

	// Get categories
	categories, err := utils.GetCategories(w, db)
	if err != nil {
		log.Printf("ERROR: Failed to get categories: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("DEBUG: Fetched categories-main")

	// We get a list of posts from the database
	posts, err := handlers.GetPosts(db, currentUser)
	if err != nil {
		errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Error retrieving posts.")
		return
	}

	// Create data for template
	data := struct {
		Posts       []models.PostView
		CurrentUser *models.User
		Categories  []models.Category
	}{
		Posts:       posts,
		CurrentUser: currentUser,
		Categories:  categories,
	}
	// Download the post creation page template
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
		"templates/header.html",
		"templates/nav.html",
		"templates/post_list.html",
		"templates/post_list_item.html",
		"templates/filters.html",
		"templates/notifications.html",
	)
	if err != nil {
		errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template not load.")
		return
	}
	tmpl.ExecuteTemplate(w, "layout", data)
}

func setupRoutes(app *App) *http.ServeMux {
	// Create WebSocket hub
	hub := handlers.NewHub()
	go hub.Run()
	mux := http.NewServeMux()

	// Home
	mux.HandleFunc("/", middleware.AuthMiddleware(app.DB, app.handler))

	// Posting
	mux.HandleFunc("/create", middleware.AuthMiddleware(app.DB, handlers.ServeFormCreatePost(app.DB)))
	mux.HandleFunc("/createin", middleware.AuthMiddleware(app.DB, handlers.HandlerCreatePost(app.DB)))
	mux.HandleFunc("/post_page/", middleware.AuthMiddleware(app.DB, handlers.ServePostByID(app.DB)))
	mux.HandleFunc("/delete_post/", middleware.AuthMiddleware(app.DB, handlers.HandlerDeletePost(app.DB)))
	mux.HandleFunc("/edit_post/", middleware.AuthMiddleware(app.DB, handlers.EditPostHandler(app.DB)))
	mux.HandleFunc("/filters_page", middleware.AuthMiddleware(app.DB, handlers.HandlePostsFilter(app.DB)))
	mux.HandleFunc("/create-comment", middleware.AuthMiddleware(app.DB, handlers.CreateCommentHandler(app.DB)))
	mux.HandleFunc("/delete_comment/", middleware.AuthMiddleware(app.DB, handlers.HandlerDeleteComment(app.DB)))
	mux.HandleFunc("/edit_comment/", middleware.AuthMiddleware(app.DB, handlers.UpdateCommentHandler(app.DB)))
	mux.HandleFunc("/like", middleware.AuthMiddleware(app.DB, handlers.HandleReaction(app.DB)))

	// Account
	mux.HandleFunc("/profile", middleware.AuthMiddleware(app.DB, handlers.HandlerProfile(app.DB)))
	mux.HandleFunc("/upload_avatar", middleware.AuthMiddleware(app.DB, handlers.UploadAvatarHandler(app.DB)))
	mux.HandleFunc("/user_page", middleware.AuthMiddleware(app.DB, handlers.HandlerUser(app.DB)))

	// Authentication
	mux.HandleFunc("/register", handlers.ServeFormRegister(app.DB))
	mux.HandleFunc("/register-submit", handlers.HandlerRegistration(app.DB))
	mux.HandleFunc("/login", handlers.ServeFormLogin(app.DB))
	mux.HandleFunc("/login-submit", handlers.HandlerLogin(app.DB))
	mux.HandleFunc("/auth/google/login", handlers.HandleGoogleLogin)
	mux.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback(app.DB))
	mux.HandleFunc("/auth/github/login", handlers.HandleGitHubLogin)
	mux.HandleFunc("/auth/github/callback", handlers.HandleGitHubCallback(app.DB))
	mux.HandleFunc("/logout", middleware.AuthMiddleware(app.DB, handlers.LogoutHandler(app.DB)))
	mux.HandleFunc("/forgot-password", handlers.ForgotPasswordHandler(app.DB))
	mux.HandleFunc("/forgot-password-submit", handlers.ForgotPasswordSubmitHandler(app.DB))
	mux.HandleFunc("/reset-password", handlers.ResetPasswordHandler(app.DB))
	mux.HandleFunc("/reset-password-submit", handlers.ResetPasswordSubmitHandler(app.DB))

	// Search
	mux.HandleFunc("/search", middleware.AuthMiddleware(app.DB, handlers.HandlerSearch(app.DB)))
	mux.HandleFunc("/profile_activity_search", middleware.AuthMiddleware(app.DB, handlers.HandlerUserActivitySearch(app.DB)))

	// Admin
	mux.HandleFunc("/admin/users", middleware.AuthMiddleware(app.DB, handlers.AdminUsersHandler(app.DB)))
	mux.HandleFunc("/admin/promote", middleware.AuthMiddleware(app.DB, handlers.PromoteHandler(app.DB)))
	mux.HandleFunc("/request-moderator", middleware.AuthMiddleware(app.DB, handlers.HandleRequestModerator(app.DB)))
	mux.HandleFunc("/admin/moderator/approve", middleware.AuthMiddleware(app.DB, handlers.ApproveModeratorRequest(app.DB)))
	mux.HandleFunc("/admin/moderator/reject", middleware.AuthMiddleware(app.DB, handlers.RejectModeratorRequest(app.DB)))
	mux.HandleFunc("/check-moderator-status", middleware.AuthMiddleware(app.DB, handlers.CheckModeratorStatusHandler(app.DB)))
	mux.HandleFunc("/admin/categories", middleware.AuthMiddleware(app.DB, handlers.AdminCategoriesPage(app.DB)))
	mux.HandleFunc("/admin/category/create", middleware.AuthMiddleware(app.DB, handlers.CreateCategoryHandler(app.DB)))
	mux.HandleFunc("/admin/categories/delete", middleware.AuthMiddleware(app.DB, handlers.DeleteCategoryHandler(app.DB)))
	mux.HandleFunc("/admin/ban", middleware.AuthMiddleware(app.DB, handlers.BanUserHandler(app.DB)))
	mux.HandleFunc("/admin/unban", middleware.AuthMiddleware(app.DB, handlers.UnbanUserHandler(app.DB)))

	// Notifications
	mux.HandleFunc("/notifications", middleware.AuthMiddleware(app.DB, handlers.HandlerGetNotifications(app.DB)))
	mux.HandleFunc("/notifications/read", middleware.AuthMiddleware(app.DB, handlers.HandlerMarkNotificationRead(app.DB)))
	mux.HandleFunc("/notifications/read-all", middleware.AuthMiddleware(app.DB, handlers.HandlerMarkAllNotificationsRead(app.DB)))
	mux.HandleFunc("/notifications/add_reply", middleware.AuthMiddleware(app.DB, handlers.HandlerAddReply(app.DB, hub)))

	return mux
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("❌ Error loading .env file:", err)
	}

	utils.LoadAndVerify()

	utils.InitOAuthConfigs()

	// Initialize database
	if err := db.SetupDatabase(); err != nil {
		log.Fatal("❌ Database setup failed:", err)
	}

	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}
	defer app.DB.Close()
	// Initialize error templates
	errors.Init("templates/error.html")

	mux := setupRoutes(app)

	// Add static file handler with correct MIME types
	mux.HandleFunc("/static/", staticFileHandler)

	// Server with middleware: RateLimit + ForceHTTPS
	handler := middleware.RateLimit(mux)

	server := &http.Server{
		Addr:         ":8080", // port for local HTTP
		Handler:      handler, // your mux with middleware
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server on port 8080
	log.Println("Server started on http://localhost:8080")
	log.Fatal(server.ListenAndServe()) // run without TLS

	// Configuring TLS with autocert for the yourforum.com domain
	// handler := middleware.RateLimit(middleware.ForceHTTPS(mux))
	// tlsConfig := security.SetupTLS("yourforum.com")
	// server := &http.Server{
	//     Addr:         ":443",
	//     Handler:      handler,
	//     TLSConfig:    tlsConfig,
	//     ReadTimeout:  5 * time.Second,
	//     WriteTimeout: 10 * time.Second,
	//     IdleTimeout:  120 * time.Second,
	// }

	// // HTTP -> HTTPS редірект
	// go func() {
	//     log.Fatal(http.ListenAndServe(":80", http.HandlerFunc(redirectToHTTPS)))
	// }()

	// fmt.Println("Server started on port 8080...")
	// log.Fatal(server.ListenAndServeTLS("", "")) // Сертифікати через autocert або інші

}
func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
}
