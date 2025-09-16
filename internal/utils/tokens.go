package utils

import "github.com/google/uuid"

func GenerateToken() string {
	return uuid.NewString()
}
