package handlers

import (
	"database/sql"
	"fmt"
	"forum/internal"
	"forum/internal/utils"
	"log"
	"net/http"
	"strconv"
)

func CreateCommentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errors.RenderError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "The HTTP method is not supported.")
			return
		}

		err := r.ParseForm()
		if err != nil {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Error parsing form data")
			return
		}

		//Get userID from context
		userID, ok := utils.MustGetUserID(w, r)
		if !ok {
			return // Return nil for both user and error
		}

		postID, err := strconv.Atoi(r.FormValue("postId"))
		if err != nil {
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid post ID.")
			return
		}

		parentCommentID, err := strconv.Atoi(r.FormValue("parentCommentId"))
		if err != nil {
			parentCommentID = 0 // or other default value for the root comment
		}

		content := r.FormValue("content")
		if content == "" {
			w.WriteHeader(http.StatusOK)
			return
		}

		commentID, err := utils.AddComment(db, postID, userID, parentCommentID, content)
		if err != nil {
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Error adding comment to database.")
			return
		}
		log.Printf("Created comment with ID: %d", commentID)
		http.Redirect(w, r, fmt.Sprintf("/post_page/%d", postID), http.StatusSeeOther)
	}
}
