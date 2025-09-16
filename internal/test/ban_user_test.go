package test

import (
	"database/sql"
	"forum/internal/handlers"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type User struct {
	ID   int
	Role string
}

func IsAdmin(u *User) bool {
	return u != nil && u.Role == "admin"
}

func IsModerator(u *User) bool {
	return u != nil && (u.Role == "moderator" || u.Role == "admin")
}

func mockDB() *sql.DB {
	db, _ := sql.Open("sqlite3", ":memory:")
	db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, role TEXT, banned INTEGER)`)
	db.Exec(`INSERT INTO users (id, role, banned) VALUES (1, 'admin', 0), (2, 'user', 0)`)
	return db
}

func TestBanUserHandler_Unauthorized(t *testing.T) {
	db := mockDB()
	handler := handlers.BanUserHandler(db) // returns http.HandlerFunc

	req := httptest.NewRequest(http.MethodPost, "/admin/ban", strings.NewReader("user_id=5"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// ❗️ Optionally, set up session/cookie if your handler expects session login

	w := httptest.NewRecorder()
	handler(w, req) // ← now correct usage

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want bool
	}{
		{"Admin user", &User{ID: 1, Role: "admin"}, true},
		{"Moderator user", &User{ID: 2, Role: "moderator"}, false},
		{"Regular user", &User{ID: 3, Role: "user"}, false},
		{"Nil user", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAdmin(tt.user); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsModerator(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want bool
	}{
		{"Admin is moderator", &User{ID: 1, Role: "admin"}, true},
		{"Moderator user", &User{ID: 2, Role: "moderator"}, true},
		{"Regular user", &User{ID: 3, Role: "user"}, false},
		{"Nil user", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModerator(tt.user); got != tt.want {
				t.Errorf("IsModerator() = %v, want %v", got, tt.want)
			}
		})
	}
}
