package handlers

import (
	"database/sql"
	// "html/template"
	"fmt"
	"forum/internal"
	"forum/internal/utils"
	"net/http"
	"strconv"
)

func UpdateCommentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errors.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed", "Method not allowed")
			return
		}

		commentIDStr := r.FormValue("commentId")
		newContent := r.FormValue("commentContent")

		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil || newContent == "" {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid data.")
			return
		}

		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil || user == nil {
			errors.RenderError(w, http.StatusUnauthorized, "Unauthorized", "Please log in.")
			return
		}

		var authorID, postID int
		err = db.QueryRow("SELECT user_id, post_id FROM comments WHERE id = ?", commentID).Scan(&authorID, &postID)
		if err != nil {
			errors.RenderError(w, http.StatusNotFound, "Not Found", "Comment not found.")
			return
		}

		//  Check permissions (admin, moderator, or comment author)
		if !utils.HasPermission(user, authorID, "edit") {
			errors.RenderError(w, http.StatusForbidden, "Forbidden", "You don't have permission to edit this comment.")
			return
		}

		_, err = db.Exec("UPDATE comments SET content = ? WHERE id = ?", newContent, commentID)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Could not update comment.")
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/post_page/%d", postID), http.StatusSeeOther)
	}
}
