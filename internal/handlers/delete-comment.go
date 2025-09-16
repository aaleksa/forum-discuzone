package handlers

import (
	"database/sql"
	"fmt"
	"forum/internal"
	"forum/internal/utils"
	"net/http"
	"strconv"
	"strings"
)

func HandlerDeleteComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil || user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get comment ID from URL
		commentIDStr := strings.TrimPrefix(r.URL.Path, "/delete_comment/")
		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid comment ID.")
			return
		}

		// Get user_id and post_id for permissions check and redirect
		var authorID, postID int
		err = db.QueryRow("SELECT user_id, post_id FROM comments WHERE id = ?", commentID).Scan(&authorID, &postID)
		if err != nil {
			errors.RenderError(w, http.StatusNotFound, "Not Found", "Comment not found.")
			return
		}

		//  Check permissions (admin, moderator, or comment author)
		if !utils.HasPermission(user, authorID, "delete") {
			errors.RenderError(w, http.StatusForbidden, "Forbidden", "You don't have permission to delete this comment.")
			return
		}

		// Delete comment
		_, err = db.Exec("DELETE FROM comments WHERE id = ?", commentID)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Error deleting comment.")
			return
		}

		// Redirect to the post page
		http.Redirect(w, r, fmt.Sprintf("/post_page/%d", postID), http.StatusSeeOther)
	}
}
