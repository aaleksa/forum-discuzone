package utils

import "forum/internal/models"

func HasPermission(user *models.User, ownerID int, action string) bool {
	switch action {
	case "edit":
		return user.Role == "admin" || user.Role == "moderator" || user.ID == ownerID
	case "delete":
		return user.Role == "admin" || user.Role == "moderator" || user.ID == ownerID
	default:
		return false
	}
}
