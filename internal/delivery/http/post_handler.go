package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"my-chi-app/internal/database/repository"
	"my-chi-app/internal/domain/entity"

	"github.com/go-chi/chi/v5"
)

// ReactionInfo is the payload response when returning reaction type information.
type ReactionInfo struct {
	ReactionTypeID int64   `json:"reaction_type_id"`
	Name           string  `json:"name"`
	Image          *string `json:"image,omitempty"`
}

// CreatePostRequest is the payload request when creating a new post.
type CreatePostRequest struct {
	Headline string  `json:"headline"`
	Text     *string `json:"text,omitempty"`
	Image    *string `json:"image,omitempty"`
}

// UpdatePostRequest is the payload request when updating a post.
type UpdatePostRequest struct {
	Headline string  `json:"headline"`
	Text     *string `json:"text,omitempty"`
	Image    *string `json:"image,omitempty"`
}

// ReactToPostRequest for reacting to a post.
type ReactToPostRequest struct {
	ReactionTypeID int64 `json:"reaction_type_id"`
}

// PostResponse is the payload response when returning post information
type PostResponse struct {
	PostID        int64         `json:"post_id"`
	Headline      string        `json:"headline"`
	Text          *string       `json:"text,omitempty"`
	Image         *string       `json:"image,omitempty"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
	IsEdited      bool          `json:"is_edited"`
	TotalReaction int64         `json:"total_reaction"`
	UserReaction  *ReactionInfo `json:"user_reaction"`
}

// @Summary Get posts by category
// @Description Fetch paginated posts from a specific category
// @Tags posts
// @Security Bearer
// @Param category_id path int true "Category ID"
// @Param limit query int false "Limit" default(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} PostResponse
// @Failure 401 {object} map[string]string
// @Router /categories/{category_id}/posts [get]
func HandleGetPostsByCategory(postRepo *repository.PostRepository, reactionRepo *repository.ReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		categoryIDStr := chi.URLParam(r, "category_id")
		categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid category_id")
			return
		}

		ctx := r.Context()

		// Paginated posts by category
		posts, err := postRepo.GetByCategory(ctx, categoryID, 1000, 0)
		if err != nil {
			InternalError(w, "failed to fetch posts")
			return
		}

		response := make([]PostResponse, len(posts))
		for i, post := range posts {
			response[i] = PostResponse{
				PostID:    post.ID,
				Headline:  post.Headline,
				Text:      post.Text,
				Image:     post.Image,
				CreatedAt: post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt: post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				IsEdited:  post.Status,
			}

			totalReactions, err := reactionRepo.CountByPost(ctx, post.ID)
			if err == nil {
				response[i].TotalReaction = totalReactions
			}

			userReaction, err := reactionRepo.GetByOwnerAndPost(ctx, userID, post.ID)
			if err == nil && userReaction != nil {
				reactionType, err := reactionTypeRepo.GetByID(ctx, userReaction.ReactionTypeID)
				if err == nil {
					response[i].UserReaction = &ReactionInfo{
						ReactionTypeID: reactionType.ID,
						Name:           reactionType.Name,
						Image:          reactionType.Image,
					}
				}
			} else if err != nil && userReaction == nil {
				response[i].UserReaction = &ReactionInfo{
					ReactionTypeID: 0,
					Name:           "",
					Image:          nil,
				}
			}
		}

		Success(w, response)
	}
}

// @Summary Get user's posts
// @Description Fetch all posts created by the authenticated user
// @Tags posts
// @Security Bearer
// @Param limit query int false "Limit" default(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} PostResponse
// @Failure 401 {object} map[string]string
// @Router /user/posts [get]
func HandleGetUserPosts(postRepo *repository.PostRepository, reactionRepo *repository.ReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		ctx := r.Context()

		posts, err := postRepo.GetByOwner(ctx, userID, 1000, 0)
		if err != nil {
			InternalError(w, "failed to fetch posts")
			return
		}

		response := buildPostResponses(ctx, posts, userID, reactionRepo, reactionTypeRepo)
		Success(w, response)
	}
}

// @Summary Get user's posts by category
// @Description Fetch all posts created by the user in a specific category
// @Tags posts
// @Security Bearer
// @Param category_id path int true "Category ID"
// @Param limit query int false "Limit" default(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} PostResponse
// @Failure 401 {object} map[string]string
// @Router /categories/{category_id}/posts/user [get]
func HandleGetUserPostsByCategory(postRepo *repository.PostRepository, reactionRepo *repository.ReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		categoryIDStr := chi.URLParam(r, "category_id")
		categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid category_id")
			return
		}

		ctx := r.Context()

		// Paginated user's posts from that category
		posts, err := postRepo.GetByOwnerAndCategory(ctx, userID, categoryID, 1000, 0)
		if err != nil {
			InternalError(w, "failed to fetch posts")
			return
		}

		response := buildPostResponses(ctx, posts, userID, reactionRepo, reactionTypeRepo)
		Success(w, response)
	}
}

// @Summary Get a post by ID
// @Description Retrieve all details including reactions details of a specific post from its ID
// @Tags posts
// @Security Bearer
// @Param post_id path int true "Post ID"
// @Success 200 {object} PostResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /posts/{post_id} [get]
func HandleGetPost(postRepo *repository.PostRepository, reactionRepo *repository.ReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		postIDStr := chi.URLParam(r, "post_id")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid post_id")
			return
		}

		ctx := r.Context()

		post, err := postRepo.GetByID(ctx, postID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "post not found")
				return
			}
			InternalError(w, "failed to fetch post")
			return
		}

		response := PostResponse{
			PostID:    post.ID,
			Headline:  post.Headline,
			Text:      post.Text,
			Image:     post.Image,
			CreatedAt: post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			IsEdited:  post.Status,
		}

		totalReactions, err := reactionRepo.CountByPost(ctx, post.ID)
		if err == nil {
			response.TotalReaction = totalReactions
		}

		userReaction, err := reactionRepo.GetByOwnerAndPost(ctx, userID, post.ID)
		if err == nil && userReaction != nil {
			reactionType, err := reactionTypeRepo.GetByID(ctx, userReaction.ReactionTypeID)
			if err == nil {
				response.UserReaction = &ReactionInfo{
					ReactionTypeID: reactionType.ID,
					Name:           reactionType.Name,
					Image:          reactionType.Image,
				}
			}
		} else if err != nil && userReaction == nil {
			response.UserReaction = &ReactionInfo{
				ReactionTypeID: 0,
				Name:           "",
				Image:          nil,
			}
		}

		Success(w, response)
	}
}

// @Summary Create a new post
// @Description Create a new post in a specific category
// @Tags posts
// @Security Bearer
// @Param category_id path int true "Category ID"
// @Param request body CreatePostRequest true "Post data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /categories/{category_id}/posts [post]
func HandleCreatePost(postRepo *repository.PostRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		categoryIDStr := chi.URLParam(r, "category_id")
		categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid category_id")
			return
		}

		var req CreatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Headline == "" {
			ValidationError(w, "headline is required")
			return
		}

		ctx := r.Context()

		post := &entity.Post{
			OwnerID:    userID,
			CategoryID: categoryID,
			Headline:   req.Headline,
			Text:       req.Text,
			Image:      req.Image,
			Status:     false,
		}

		post, err = postRepo.Create(ctx, post)
		if err != nil {
			InternalError(w, "failed to create post")
			return
		}

		Success(w, MessageResponse{
			Message: "Post created successfully!",
		})
	}
}

// @Summary Update a post
// @Description Update the content of a post
// @Tags posts
// @Security Bearer
// @Param post_id path int true "Post ID"
// @Param request body UpdatePostRequest true "Updated post data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /posts/{post_id} [put]
func HandleUpdatePost(postRepo *repository.PostRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		postIDStr := chi.URLParam(r, "post_id")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid post_id")
			return
		}

		var req UpdatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Headline == "" {
			ValidationError(w, "headline is required")
			return
		}

		ctx := r.Context()

		post, err := postRepo.GetByID(ctx, postID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "post not found")
				return
			}
			InternalError(w, "failed to fetch post")
			return
		}

		if post.OwnerID != userID {
			Forbidden(w, "you can only update your own posts")
			return
		}

		post.Headline = req.Headline
		post.Text = req.Text
		post.Image = req.Image

		if err := postRepo.Update(ctx, post); err != nil {
			InternalError(w, "failed to update post")
			return
		}

		Success(w, MessageResponse{
			Message: "Post updated successfully!",
		})
	}
}

// @Summary Delete a post
// @Description Delete a post and its associated data
// @Tags posts
// @Security Bearer
// @Param post_id path int true "Post ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /posts/{post_id} [delete]
func HandleDeletePost(postRepo *repository.PostRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		postIDStr := chi.URLParam(r, "post_id")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid post_id")
			return
		}

		ctx := r.Context()

		post, err := postRepo.GetByID(ctx, postID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "post not found")
				return
			}
			InternalError(w, "failed to fetch post")
			return
		}

		if post.OwnerID != userID {
			Forbidden(w, "you can only delete your own posts")
			return
		}

		if err := postRepo.Delete(ctx, postID); err != nil {
			InternalError(w, "failed to delete post")
			return
		}

		Success(w, MessageResponse{
			Message: "Post deleted successfully!",
		})
	}
}

// @Summary React to a post
// @Description Add or update a reaction to a specific post
// @Tags posts
// @Security Bearer
// @Param post_id path int true "Post ID"
// @Param request body ReactToPostRequest true "Reaction type"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /posts/{post_id}/react [post]
func HandleReactToPost(reactionRepo *repository.ReactionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		postIDStr := chi.URLParam(r, "post_id")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid post_id")
			return
		}

		var req ReactToPostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.ReactionTypeID == 0 {
			ValidationError(w, "reaction_type_id is required")
			return
		}

		ctx := r.Context()

		reaction := &entity.Reaction{
			PostID:         postID,
			OwnerID:        userID,
			ReactionTypeID: req.ReactionTypeID,
		}

		_, err = reactionRepo.Upsert(ctx, reaction)
		if err != nil {
			InternalError(w, "failed to record reaction")
			return
		}

		Success(w, MessageResponse{
			Message: "Reaction recorded!",
		})
	}
}

// buildPostResponses converts post entities to PostResponse with reaction details
// Includes total reactions and user's reaction
func buildPostResponses(ctx context.Context, posts []*entity.Post, userID int64, reactionRepo *repository.ReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) []PostResponse {
	response := make([]PostResponse, len(posts))
	for i, post := range posts {
		response[i] = PostResponse{
			PostID:    post.ID,
			Headline:  post.Headline,
			Text:      post.Text,
			Image:     post.Image,
			CreatedAt: post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			IsEdited:  post.Status,
		}

		totalReactions, err := reactionRepo.CountByPost(ctx, post.ID)
		if err == nil {
			response[i].TotalReaction = totalReactions
		}

		userReaction, err := reactionRepo.GetByOwnerAndPost(ctx, userID, post.ID)
		if err == nil && userReaction != nil {
			reactionType, err := reactionTypeRepo.GetByID(ctx, userReaction.ReactionTypeID)
			if err == nil {
				response[i].UserReaction = &ReactionInfo{
					ReactionTypeID: reactionType.ID,
					Name:           reactionType.Name,
				}
			}
		}
	}
	return response
}
