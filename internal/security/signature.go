package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"strings"
)

var hmacSecret []byte

// InitHMACSecret must be called during app startup
func InitHMACSecret() {
	secret := os.Getenv("SESSION_HMAC_SECRET")
	if secret == "" {
		log.Fatal("❌ SESSION_HMAC_SECRET is not set")
	}
	hmacSecret = []byte(secret)
}

func SignSessionID(sessionID string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(sessionID))
	signature := computeHMAC(encoded)
	return encoded + "|" + signature
}

func VerifySignedSessionID(signed string) (string, bool) {
	parts := splitSigned(signed)
	if len(parts) != 2 {
		return "", false
	}
	encoded, sig := parts[0], parts[1]

	if computeHMAC(encoded) != sig {
		return "", false
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", false
	}

	return string(data), true
}

func computeHMAC(message string) string {
	h := hmac.New(sha256.New, hmacSecret)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func splitSigned(s string) []string {
	return strings.SplitN(s, "|", 2)
}
