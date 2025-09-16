package handlers

import (
	"fmt"

	"database/sql"
	"forum/internal"
	"forum/internal/utils"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

func ForgotPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(
			"templates/layout_auth.html",
			"templates/forgot_password.html",
			"templates/header.html",
		)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template not loaded.")
			return
		}

		err = tmpl.ExecuteTemplate(w, "layout", nil)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template execution error.")
		}
	}
}

func ForgotPasswordSubmitHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")

		var userID int
		err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
		if err != nil {
			errors.RenderError(w, http.StatusNotFound, "Not Found", "User not found.")
			return
		}

		token := utils.GenerateToken()
		expiration := time.Now().Add(3 * time.Hour)

		_, err = db.Exec(`INSERT INTO password_resets (user_id, token, expires_at) VALUES (?, ?, ?)`, userID, token, expiration)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Something went wrong.")
			return
		}

		err = sendResetEmail(email, token)
		if err != nil {
			log.Printf("Send email error: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Failed to send reset email.")
			return
		}

		// ✅ Return 200 OK for AJAX
		w.WriteHeader(http.StatusOK)

		// // For testing, just print reset link
		// fmt.Fprintf(w, "Reset link: http://localhost:8080/reset-password?token=%s", token)
	}
}

func sendResetEmail(email, token string) error {
	// 0. Пconfiguration check
	requiredVars := []string{"EMAIL_FROM_NAME", "EMAIL_FROM_ADDRESS", "SENDGRID_API_KEY"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			return fmt.Errorf("%s environment variable not set", v)
		}
	}

	// 1.Email validation
	if !isValidEmail(email) {
		return fmt.Errorf("invalid email address: %s", email)
	}

	// 2. Sender settings
	from := mail.NewEmail(
		os.Getenv("EMAIL_FROM_NAME"),    // "DiscuZone Forum"
		os.Getenv("EMAIL_FROM_ADDRESS"), // Verified address in SendGrid
	)

	// 3. Link generation
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default для development
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

	// 4.HTML content (optimized)
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
    <html>
    <head><meta charset="UTF-8"></head>
    <body style="font-family: Arial, sans-serif;">
        <h2 style="color: #333;">Password Reset</h2>
        <p>You requested a password reset. Click the button below:</p>
        <a href="%s" style="%s">Reset Password</a>
        <p><small>Link expires in 1 hour. If you didn't request this, please ignore this email.</small></p>
    </body>
    </html>`, resetLink, "background:#4CAF50; color:white; padding:10px 20px; text-decoration:none; border-radius:5px;")

	// 5. Sending a letter
	message := mail.NewSingleEmail(from, "Password Reset Request", mail.NewEmail("", email),
		fmt.Sprintf("Password Reset Link: %s", resetLink), htmlContent)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)

	if err != nil {
		log.Printf("SendGrid error: %v, Status: %d", err, response.StatusCode)
		log.Printf("SendGrid response body: %s", response.Body)
		return fmt.Errorf("failed to send email")
	}

	log.Printf("Email sent to %s (Status: %d)", email, response.StatusCode)
	return nil
}

// Helper function for email validation
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
