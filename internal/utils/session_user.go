package utils

import (
	"database/sql"
	"fmt"
	"forum/internal/models"

	"forum/internal/security"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Function to get user data
func GetUserFromSession(w http.ResponseWriter, r *http.Request, db *sql.DB) (*models.User, error) {
	// Get userID from context
	userID, ok := MustGetUserID(w, r)
	if !ok {
		return nil, nil // Error already handled, exit
	}

	// Get user data
	var user models.User
	var createdAt time.Time
	var avatarURL sql.NullString // Use NullString to handle NULL
	err := db.QueryRow(`
        SELECT id, username, email, created_at, role, avatar_url 
        FROM users 
        WHERE id = ?
    `, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&createdAt,
		&user.Role,
		&avatarURL,
	)
	if err != nil {
		return nil, fmt.Errorf("user query error: %v", err)
	}

	// Handle NULL values for avatar_url
	if avatarURL.Valid {
		user.AvatarPath = avatarURL.String
	} else {
		user.AvatarPath = "" // or default value "/default-avatar.png"
	}

	user.CreatedAt = FormatDate(createdAt)

	// Get additional data
	user.CreatedPosts, err = GetCreatedPosts(db, user.ID)
	if err != nil {
		log.Printf("Warning: failed to get created posts: %v", err)
	}

	user.LikedPosts, err = GetLikedPosts(db, user.ID)
	if err != nil {
		log.Printf("Warning: failed to get liked posts: %v", err)
	}

	user.DislikePosts, err = GetDislikes(db, user.ID)
	if err != nil {
		log.Printf("Warning: failed to get dislikes: %v", err)
	}
	log.Printf("User in GetUserFromSession: %v", user)
	return &user, nil
}

// getAllUsers gets all users from the database
func GetAllUsers(db *sql.DB) ([]models.User, error) {
	rows, err := db.Query(`
		SELECT id, username, email, password, created_at, role, banned
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.Role, &u.Banned)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func CheckUserExists(db *sql.DB, email string) (bool, models.NewUser, error) {
	var user models.NewUser

	query := "SELECT id, username, email, password, created_at, role, banned FROM users WHERE TRIM(email) = ?"
	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.Role,
		&user.Banned,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, models.NewUser{}, nil
		}
		return false, models.NewUser{}, err
	}
	return true, user, nil
}

func ValidateAndLoginUser(w http.ResponseWriter, r *http.Request, db *sql.DB, creds LoginCredentials, oauthMarker bool) {
	// Валідатор користувача за email
	user, err := ValidateUserForLogin(db, creds.Email)
	if err != nil {
		switch err.Error() {
		case "invalid credentials":
			RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		case "account banned":
			RespondWithJSON(w, http.StatusForbidden, map[string]interface{}{
				"error":   "Account banned",
				"message": "Your account has been blocked.",
				"contact": "support@example.com",
			})
		default:
			RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	// Suppose the user model has an OauthMarker bool field
	if !oauthMarker { // If this is NOT an OAuth user, check the password
		if !security.CheckPasswordHash(creds.Password, user.PasswordHash) {
			RespondWithError(w, http.StatusUnauthorized, "Invalid email or password check")
			return
		}
	} else {
		// If OAuth user, skip password verification
		log.Printf("OAuth user detected, skip password verification for email: %s", user.Email)
	}
	log.Printf("SessionValidateAndLoginUser user %d:", user.ID)
	// Create a session
	if err := security.CreateSession(w, user.ID, db); err != nil {
		log.Printf("Session creation error for user %d: %v", user.ID, err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	if oauthMarker {
		// ⬇️ we use DelayedRedirect
		DelayedRedirect(w, "/user_page", 300)
		return

	} else {
		// Successful response (JSON with redirect)
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message":  "Login successful",
			"redirect": "/user_page",
			"user": map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
			},
		})
	}

}

// Function for general user validation before login
func ValidateUserForLogin(db *sql.DB, email string) (*models.NewUser, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Check for user existence
	exists, user, err := CheckUserExists(db, email)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	if !exists {
		time.Sleep(500 * time.Millisecond) // Delay for safety
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check ban
	if user.Banned {
		return nil, fmt.Errorf("account banned")
	}

	return &user, nil
}

// DelayedRedirect returns HTML with a modal notification and a delayed redirect via JS
func DelayedRedirect(w http.ResponseWriter, target string, delayMs int) {
	if delayMs <= 0 {
		delayMs = 3000 // increased default delay to allow reading the modal
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Redirecting...</title>
            <style>
                .modal {
                    display: block;
                    position: fixed;
                    z-index: 100;
                    left: 0;
                    top: 0;
                    width: 100%%;
                    height: 100%%;
                    background-color: rgba(0,0,0,0.7);
                }
                .modal-content {
                    background-color: #fefefe;
                    margin: 15%% auto;
                    padding: 20px;
                    border: 1px solid #888;
                    width: 80%%;
                    max-width: 500px;
                    border-radius: 5px;
                    box-shadow: 0 4px 8px 0 rgba(0,0,0,0.2);
                }
                .close-btn {
                    color: #aaa;
                    float: right;
                    font-size: 28px;
                    font-weight: bold;
                    cursor: pointer;
                }
                .close-btn:hover {
                    color: black;
                }
                .modal-footer {
                    margin-top: 20px;
                    text-align: right;
                }
                .modal-footer button {
                    padding: 8px 16px;
                    margin-left: 10px;
                    cursor: pointer;
                    border: none;
                    border-radius: 4px;
                }
                .contact-btn {
                    background-color: #f44336;
                    color: white;
                }
            </style>
        </head>
        <body>
            <!-- Success Modal -->
            <div id="successModal" class="modal">
                <div class="modal-content">
                    <h2>✅ Success!</h2>
                    <p>Login successful! You will be redirected shortly...</p>
                    <div class="modal-footer">
                        <button onclick="proceedRedirect()">Continue now</button>
                    </div>
                </div>
            </div>

            <script>
                function proceedRedirect() {
                    window.location.href = %q;
                }
                
                // Auto-redirect after delay
                setTimeout(() => {
                    proceedRedirect();
                }, %d);
            </script>
        </body>
        </html>
    `, target, delayMs)
}
