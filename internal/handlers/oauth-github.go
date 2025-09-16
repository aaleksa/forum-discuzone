package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/internal/utils"
	"log"
	"net/http"
)

// Redirect the user to GitHub for authorization
func HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := utils.GitHubOAuthConfig.AuthCodeURL("state-token")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGitHubCallback handles the response from GitHub OAuth
func HandleGitHubCallback(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Check state
		if r.FormValue("state") != "state-token" {
			http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
			return
		}

		// 2. GitHub error handling
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errorDesc := r.URL.Query().Get("error_description")
			log.Printf("[GitHub OAuth] Error from GitHub: %s - %s\n", errMsg, errorDesc)
			handleOAuthError(w, r, "GitHub authentication failed: "+errorDesc, http.StatusUnauthorized)
			return
		}

		// 3. Get authorization code
		code := r.FormValue("code")
		if code == "" {
			log.Println("[GitHub OAuth] Missing authorization code")
			handleOAuthError(w, r, "Missing authorization code", http.StatusBadRequest)
			return
		}

		// 4. Exchange code for token
		log.Println("[GitHub OAuth] Exchanging code for token")
		token, err := utils.GitHubOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("[GitHub OAuth] Token exchange failed: %v\n", err)
			handleOAuthError(w, r, "Failed to authenticate with GitHub", http.StatusInternalServerError)
			return
		}

		// 5. Get user profile
		log.Println("[GitHub OAuth] Fetching user info")
		client := utils.GitHubOAuthConfig.Client(context.Background(), token)

		// Get main profile info
		profileResp, err := client.Get("https://api.github.com/user")
		if err != nil {
			handleOAuthError(w, r, "Failed to get GitHub user info", http.StatusInternalServerError)
			return
		}
		defer profileResp.Body.Close()

		var profile struct {
			ID        int64  `json:"id"`
			Login     string `json:"login"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
		}
		if err := json.NewDecoder(profileResp.Body).Decode(&profile); err != nil {
			handleOAuthError(w, r, "Failed to decode GitHub user info", http.StatusInternalServerError)
			return
		}

		// Get email
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			handleOAuthError(w, r, "Failed to get GitHub email", http.StatusInternalServerError)
			return
		}
		defer emailResp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil || len(emails) == 0 {
			handleOAuthError(w, r, "Failed to parse GitHub email response", http.StatusInternalServerError)
			return
		}

		var email string
		for _, e := range emails {
			if e.Primary && e.Verified {
				email = e.Email
				break
			}
		}
		if email == "" {
			handleOAuthError(w, r, "No verified primary email found", http.StatusUnauthorized)
			return
		}

		userID := fmt.Sprintf("%d", profile.ID)
		provider := "github"

		log.Printf("[GitHub OAuth] User info received: %s (%s)\n", email, userID)

		// 6. Check if user exists
		if IsUserExists(db, "", email, provider, userID) {
			log.Printf("[GitHub OAuth] User already exists: %s\n", email)
			credentials := utils.LoginCredentials{
				Email:    email,
				Password: "",
			}
			utils.ValidateAndLoginUser(w, r, db, credentials, true)
			return
		}

		// 7. Register new user
		username := generateUsername(profile.Name, email)
		avatar := profile.AvatarURL
		if avatar == "" {
			avatar = "/static/images/default-avatar.png"
		}

		log.Printf("[GitHub OAuth] Registering new user: %s (%s)\n", username, email)
		err = СreateOAuthUser(db, username, email, provider, userID, avatar)
		if err != nil {
			log.Printf("[GitHub OAuth] Failed to create user: %v\n", err)
			handleOAuthError(w, r, "Failed to create user account", http.StatusInternalServerError)
			return
		}

		// 8. Login newly registered user
		credentials := utils.LoginCredentials{
			Email:    email,
			Password: "",
		}
		utils.ValidateAndLoginUser(w, r, db, credentials, true)
		log.Printf("[GitHub OAuth] User successfully registered and logged in: %s\n", email)
	}
}
