package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/internal/utils"
	"log"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

// HandleGoogleLogin redirects the user to Google OAuth
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Redirect to Google OAuth
	url := utils.GoogleOAuthConfig.AuthCodeURL("state-token")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback handles the response from Google OAuth
func HandleGoogleCallback(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Logging the start of processing
		if r.FormValue("state") != "state-token" {
			http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
			return
		}

		// 2. Google error handling
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errorDesc := r.URL.Query().Get("error_description")
			log.Printf("[Google OAuth] Error from Google: %s - %s\n", errMsg, errorDesc)
			handleOAuthError(w, r, "Google authentication failed: "+errorDesc, http.StatusUnauthorized)
			return
		}

		// 3. Obtaining the authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			log.Println("[Google OAuth] Missing authorization code")
			handleOAuthError(w, r, "Missing authorization code", http.StatusBadRequest)
			return
		}

		// 4. Exchange code for token
		log.Println("[Google OAuth] Exchanging code for token")
		token, err := utils.GoogleOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("[Google OAuth] Token exchange failed: %v\n", err)
			handleOAuthError(w, r, "Failed to authenticate with Google", http.StatusInternalServerError)
			return
		}

		// 5. Getting user information
		log.Println("[Google OAuth] Fetching user info")
		userInfo, err := getGoogleUserInfo(token.AccessToken)
		if err != nil {
			log.Printf("[Google OAuth] Failed to get user info: %v\n", err)
			handleOAuthError(w, r, "Failed to get user information", http.StatusInternalServerError)
			return
		}

		log.Printf("[Google OAuth] User info received: %s (%s)\n", userInfo.Email, userInfo.Sub)

		// 6. Checking for user existence
		if IsUserExists(db, "", userInfo.Email, "google", userInfo.Sub) {
			log.Printf("[Google OAuth] User already exists: %s\n", userInfo.Email)

			// Authorize an existing user
			credentials := utils.LoginCredentials{
				Email:    userInfo.Email,
				Password: "",
			}
			utils.ValidateAndLoginUser(w, r, db, credentials, true)
			return
		}

		// 7. Registering a new user
		username := generateUsername(userInfo.Name, userInfo.Email)
		avatarURL := userInfo.Picture
		if avatarURL == "" {
			avatarURL = "/static/images/default-avatar.png"
		}

		log.Printf("[Google OAuth] Registering new user: %s (%s)\n", username, userInfo.Email)
		err = СreateOAuthUser(db, username, userInfo.Email, "google", userInfo.Sub, avatarURL)
		if err != nil {
			log.Printf("[Google OAuth] Failed to create user: %v\n", err)
			handleOAuthError(w, r, "Failed to create user account", http.StatusInternalServerError)
			return
		}

		// 8. Authorization of a new user
		credentials := utils.LoginCredentials{
			Email:    userInfo.Email,
			Password: "",
		}
		utils.ValidateAndLoginUser(w, r, db, credentials, true)
		log.Printf("[Google OAuth] User successfully registered and logged in: %s\n", userInfo.Email)
	}
}

// getGoogleUserInfo gets user information from Google API
func getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google API returned status: %s", resp.Status)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	if !userInfo.EmailVerified {
		return nil, fmt.Errorf("email %s is not verified", userInfo.Email)
	}

	return &userInfo, nil
}

// handleOAuthError handles OAuth errors
func handleOAuthError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	log.Printf("[OAuth Error] %s\n", message)

	if strings.Contains(strings.ToLower(r.Header.Get("Accept")), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{
			"error": message,
		})
	} else {
		http.Redirect(w, r, "/login?error="+url.QueryEscape(message), http.StatusSeeOther)
	}
}

// generateUsername generates a username
func generateUsername(name, email string) string {
	if name != "" {
		// Name normalization
		username := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return unicode.ToLower(r)
			}
			if unicode.IsSpace(r) {
				return '_'
			}
			return -1
		}, name)

		// Removing extra underlines
		username = strings.ReplaceAll(username, "__", "_")
		username = strings.Trim(username, "_")

		if username != "" {
			return username
		}
	}

	// Using email as a fallback
	localPart := strings.Split(email, "@")[0]
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return '_'
	}, localPart)
}

// GoogleUserInfo contains user information from Google
type GoogleUserInfo struct {
	Sub           string `json:"sub"`            // Unique identifier
	Email         string `json:"email"`          // Email address
	EmailVerified bool   `json:"email_verified"` // Whether email is verified
	Name          string `json:"name"`           // Full name
	GivenName     string `json:"given_name"`     // First name
	FamilyName    string `json:"family_name"`    // Last name
	Picture       string `json:"picture"`        // Avatar URL
	Locale        string `json:"locale"`         // Language/locale
}
