package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// List of required env variables
var requiredKeys = []string{
	"GOOGLE_CLIENT_ID",
	"GOOGLE_CLIENT_SECRET",
	"GOOGLE_REDIRECT_URL",
	"GITHUB_CLIENT_ID",
	"GITHUB_CLIENT_SECRET",
	"GITHUB_REDIRECT_URL",
}

// LoadAndVerify loads the .env file and checks required variables
func LoadAndVerify() {
	// Get absolute path to .env
	envPath, err := filepath.Abs(".env")
	if err != nil {
		log.Fatalf("❌ Could not resolve absolute path to .env file: %v", err)
	}

	// Load .env
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("❌ Error loading .env file from %s: %v", envPath, err)
	}
	log.Println("✅ .env file loaded successfully from:", envPath)

	// Print current working directory for debugging
	if cwd, err := os.Getwd(); err == nil {
		log.Println("📁 Working directory:", cwd)
	}

	// Check required environment variables
	for _, key := range requiredKeys {
		val := os.Getenv(key)
		if val == "" {
			log.Printf("⚠️  WARNING: %s is empty or not set", key)
		} else {
			log.Printf("🔑 %s: %s", key, val)
		}
	}
}
