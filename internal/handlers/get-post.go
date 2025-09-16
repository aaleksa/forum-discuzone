package handlers

import (
	"database/sql"
	errors "forum/internal"
	"forum/internal/models"
	"forum/internal/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func ServePostByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/post_page/")
		// log.Printf("[DEBUG] ServePostByID called with URL path: %s", r.URL.Path)

		postID, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("[ERROR] Post not found: ID=%d, err=%v", postID, err)
			errors.RenderError(w, http.StatusBadRequest, "Bad Request", "Invalid post ID.")
			return
		}

		post, err := utils.GetPostByID(db, postID)
		if err != nil {
			log.Printf("[ERROR] Post not found: ID=%d, err=%v", postID, err)
			errors.RenderError(w, http.StatusNotFound, "Not Found", "Post not found.")
			return
		}
		// log.Printf("[DEBUG] Post found: ID=%d, Title=%s", post.ID, post.Title)

		// Get comments with debug output
		comments, err := utils.GetCommentsByPostID(db, postID)
		if err != nil {
			log.Printf("[ERROR] Failed to fetch comments for post %d: %v", postID, err)
			http.Error(w, "Error fetching comments", http.StatusInternalServerError)
			return
		}
		// log.Printf("[DEBUG] %d comments fetched for post %d", len(comments), postID)

		// Debug: Print number of comments found
		// for i, c := range comments {
		// 	log.Printf("[DEBUG] Comment %d: ID=%d, UserID=%d, UserName=%s, Content=%s, Likes=%d, Dislikes=%d, CreatedAt=%s",
		// 		i+1, c.ID, c.UserID, c.UserName, c.Content, c.Likes, c.Dislikes, c.CreatedAt)
		// }

		// // Debug: Print each comment
		// for i, comment := range comments {
		// 	log.Printf("[DEBUG] Comment %d: ID=%d, UserName=%s, Content=%s", i+1, comment.ID, comment.UserName, comment.Content)
		// }

		// Try to get user but don't fail if not logged in
		user, err := utils.GetUserFromSession(w, r, db)
		if err != nil {
			log.Printf("[WARN] Session check warning (guest access?): %v", err)
			user = nil
		}

		var canModifyPost bool
		canModifyComments := make(map[int]bool)

		if user != nil {
			canModifyPost = utils.HasPermission(user, post.UserID, "edit")
			for _, c := range comments {
				canModifyComments[c.ID] = utils.HasPermission(user, c.UserID, "edit")
			}
		}

		// Initialize variables for reactions
		var postReaction string
		commentReactions := make(map[int]string)

		if user != nil {
			postReaction, err = getUserReaction(db, user.ID, "post", postID)
			if err != nil {
				log.Printf("[ERROR] Failed to get user reaction for post %d: %v", postID, err)
				http.Error(w, "Error getting user reaction for post", http.StatusInternalServerError)
				return
			}

			for _, comment := range comments {
				reaction, err := getUserReaction(db, user.ID, "comment", comment.ID)
				if err != nil {
					log.Printf("[ERROR] Failed to get user reaction for comment %d: %v", comment.ID, err)
					errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Error getting user reaction for comment.")
					return
				}
				commentReactions[comment.ID] = reaction
			}
		}

		// Get replies
		repliesMap, err := utils.GetRepliesForComments(db, postID)
		if err != nil {
			log.Printf("[ERROR] Failed to fetch replies for post %d: %v", postID, err)
			http.Error(w, "Error getting replies", http.StatusInternalServerError)
			return
		}
		// log.Printf("[DEBUG] Replies fetched for post %d: %d", postID, len(repliesMap))

		// Creating a reaction slice for a template
		var commentsWithReactions []models.CommentWithReaction
		for _, c := range comments {
			reaction := commentReactions[c.ID]
			commentsWithReactions = append(commentsWithReactions, models.CommentWithReaction{
				Comment:      c,
				UserReaction: reaction,
			})
		}
		// log.Printf("[DEBUG] commentsWithReactions slice created, length=%d", len(commentsWithReactions))

		tags, err := utils.GetPostTags(db, postID)
		if err != nil {
			log.Printf("[WARN] Failed to get tags for post %d: %v", postID, err)
			tags = []string{} // Set empty slice if error
		}

		data := models.PostPageData{
			Post:             post,
			Comments:         commentsWithReactions,
			CurrentUser:      user,
			UserReaction:     postReaction,
			CommentReactions: commentReactions,
			Tags:             tags,
			CanModifyPost:    canModifyPost,
			CanModifyComment: canModifyComments,
			Replies:          repliesMap,
		}

		// Debug: Print final data
		log.Printf("[DEBUG] Final PostPageData prepared: PostID=%d, Comments=%d, Replies=%d", data.Post.ID, len(data.Comments), len(data.Replies))

		tmpl, err := template.ParseFiles(
			"templates/layout.html",
			"templates/post_page.html",
			"templates/header.html",
			"templates/nav.html",
			"templates/post_item.html",
			"templates/comment.html",
			"templates/add_comment.html",
			"templates/notifications.html",
		)
		if err != nil {
			log.Printf("[ERROR] Template parsing failed: %v", err)
			errors.RenderError(w, http.StatusInternalServerError, "Internal Server Error", "Template error.")
			return
		}

		err = tmpl.ExecuteTemplate(w, "layout", data)
		if err != nil {
			log.Printf("[ERROR] Template execution failed: %v", err)
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
	}
}
