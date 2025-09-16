package utils

import (
	"errors"
	"log"
	"net/http"
)

type contextKey string

const (
	// UserIDKey is the key used to store and retrieve user ID from request context
	UserIDKey contextKey = "userID"
)

// GetUserIDFromContext retrieves user ID from request context
// Returns:
//   - int: user ID if found
//   - error: ErrNoUserInContext if user ID not found or has wrong type
func GetUserIDFromContext(r *http.Request) (int, error) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		return 0, ErrNoUserInContext
	}
	return userID, nil
}

// MustGetUserID retrieves user ID from context or writes error response
// Returns:
//   - int: user ID if found
//   - bool: false if user ID not found (error response already written)
func MustGetUserID(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		return 0, false
	}
	log.Printf("MustGetUserID in MustGetUserID: %v", userID)
	return userID, true
}

// Errors
var (
	// ErrNoUserInContext indicates no user ID was found in the request context
	ErrNoUserInContext = errors.New("no user ID found in request context")
)
