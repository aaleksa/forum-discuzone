package utils

import (
	"errors"
	"math/rand"
	"time"
)

func GenerateRandomString(length int, options ...string) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}

	const (
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		numbers   = "0123456789"
		symbols   = "-_=+!@#$%^&*()[]{}|;:,.<>?"
	)

	var charset string

	// If no option is passed - use all characters
	if len(options) == 0 {
		charset = uppercase + lowercase + numbers + symbols
	} else {
		for _, opt := range options {
			switch opt {
			case "uppercase":
				charset += uppercase
			case "lowercase":
				charset += lowercase
			case "numbers":
				charset += numbers
			case "symbols":
				charset += symbols
			}
		}
	}

	if charset == "" {
		return "", errors.New("no character set selected")
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result), nil
}
