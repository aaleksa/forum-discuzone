package test

import (
	"forum/internal/middleware"
	"forum/internal/security"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	password := "secure password 123"
	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("Hashing error: %v", err)
	}

	if !security.CheckPasswordHash(password, hash) {
		t.Error("Password verification error")
	}

	if security.CheckPasswordHash("incorrect password", hash) {
		t.Error("An incorrect password was entered.")
	}
}

func TestSessionUUID(t *testing.T) {
	session1 := uuid.New().String()
	session2 := uuid.New().String()

	if session1 == session2 {
		t.Error("UUIDs must be unique")
	}

	if len(session1) != 36 { // UUIDv4 length
		t.Error("Invalid UUID length")
	}
}

// Simple handler for the test
func testHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestRateLimit(t *testing.T) {
	handler := middleware.RateLimit(http.HandlerFunc(testHandler))

	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{}

	// Use fixed IP via header
	ip := "127.0.0.1"

	for i := 0; i < 30; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header.Set("X-Real-IP", ip)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i+1, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK at request %d, got %d", i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// 31st request - should return 429
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("X-Real-IP", ip)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("31st request failed: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestSignAndVerifySessionID(t *testing.T) {
	// Initialize the secret (for testing)
	os.Setenv("SESSION_HMAC_SECRET", "test-secret-key")
	security.InitHMACSecret()

	testSessionID := "session123"

	// Sign the sessionID
	signed := security.SignSessionID(testSessionID)
	if signed == "" {
		t.Fatal("SignSessionID returned empty string")
	}

	// Check the signed sessionID
	decoded, ok := security.VerifySignedSessionID(signed)
	if !ok {
		t.Fatalf("VerifySignedSessionID failed to verify valid signature: %s", signed)
	}
	if decoded != testSessionID {
		t.Errorf("Expected decoded sessionID to be %q, got %q", testSessionID, decoded)
	}
}

func TestVerifySignedSessionID_Invalid(t *testing.T) {
	os.Setenv("SESSION_HMAC_SECRET", "test-secret-key")
	security.InitHMACSecret()

	// Incorrect format (no '|')
	invalid := "invalidsignaturestring"
	_, ok := security.VerifySignedSessionID(invalid)
	if ok {
		t.Error("VerifySignedSessionID should return false for invalid input")
	}

	// Fake signature
	original := security.SignSessionID("session123")
	fake := original + "tampered"
	_, ok = security.VerifySignedSessionID(fake)
	if ok {
		t.Error("VerifySignedSessionID should return false for tampered signature")
	}
}
