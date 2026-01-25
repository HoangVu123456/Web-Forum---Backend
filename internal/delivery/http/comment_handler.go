package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"my-chi-app/internal/database/repository"
	"my-chi-app/internal/domain/entity"
)

// CreateCommentRequest is the payload for creating a new comment or reply
type CreateCommentRequest struct {
	Text  *string `json:"text"`
	Image *string `json:"image"`
}

// UpdateCommentRequest is the payload for updating a comment or reply
type UpdateCommentRequest struct {
	Text  *string `json:"text"`
	Image *string `json:"image"`
}

// ReactionCommentRequest is the payload for the reaction to a comment or reply
type ReactCommentRequest struct {
	ReactionTypeID int64 `json:"reaction_type_id"`
}

// CommentResponse is the response shape for comments and replies
type CommentResponse struct {
	CommentID            int64         `json:"comment_id"`
	CommentOwnerUsername string        `json:"comment_owner_username"`
	ProfilePicture       *string       `json:"comment_owner_profile_picture"`
	Text                 string        `json:"text"`
	Image                *string       `json:"image"`
	CreatedAt            string        `json:"created_at"`
	UpdatedAt            string        `json:"updated_at"`
	IsEdited             bool          `json:"is_edited"`
	TotalReaction        int64         `json:"total_reaction"`
	UserReaction         *ReactionInfo `json:"user_reaction"`
}

// @Summary Get comments by post
// @Description Fetch paginated comments to a specific post
// @Tags comments
// @Security Bearer
// @Param post_id path int true "Post ID"
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} CommentResponse
// @Failure 401 {object} map[string]string
// @Router /posts/{post_id}/comments [get]
func HandleGetCommentsByPost(commentRepo *repository.CommentRepository, userRepo *repository.UserRepository, commentReactionRepo *repository.CommentReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository, postRepo *repository.PostRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		postIDStr := chi.URLParam(r, "post_id")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid post_id")
			return
		}

		_, err = postRepo.GetByID(r.Context(), postID)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "post not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}

		// Pagination
		limit, offset := int32(1000), int32(0)
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.ParseInt(l, 10, 32); err == nil {
				limit = int32(v)
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if v, err := strconv.ParseInt(o, 10, 32); err == nil {
				offset = int32(v)
			}
		}

		comments, err := commentRepo.ListByPost(r.Context(), postID, limit, offset)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		responses, err := buildCommentResponses(r.Context(), comments, userID, userRepo, commentReactionRepo, reactionTypeRepo)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Success(w, responses)
	}
}

// @Summary Get replies to a comment
// @Description Fetch paginated replies to a specific comment
// @Tags comments
// @Security Bearer
// @Param comment_id path int true "Comment ID"
// @Param limit query int false "Limit" default(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} CommentResponse
// @Failure 401 {object} map[string]string
// @Router /comments/{comment_id}/replies [get]
func HandleGetRepliesByComment(commentRepo *repository.CommentRepository, userRepo *repository.UserRepository, commentReactionRepo *repository.CommentReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		commentIDStr := chi.URLParam(r, "comment_id")
		commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid comment_id")
			return
		}

		_, err = commentRepo.GetByID(r.Context(), commentID)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "comment not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}

		// Pagination
		limit, offset := int32(1000), int32(0)
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.ParseInt(l, 10, 32); err == nil {
				limit = int32(v)
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if v, err := strconv.ParseInt(o, 10, 32); err == nil {
				offset = int32(v)
			}
		}

		replies, err := commentRepo.ListByParent(r.Context(), commentID, limit, offset)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		responses, err := buildCommentResponses(r.Context(), replies, userID, userRepo, commentReactionRepo, reactionTypeRepo)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Success(w, responses)
	}
}

// @Summary Get user's comments
// @Description Fetch paginated comments of the authenticated user
// @Tags comments
// @Security Bearer
// @Param limit query int false "Limit" default(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} CommentResponse
// @Failure 401 {object} map[string]string
// @Router /user/comments [get]
func HandleGetUserComments(commentRepo *repository.CommentRepository, userRepo *repository.UserRepository, commentReactionRepo *repository.CommentReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		// Pagination
		limit, offset := int32(1000), int32(0)
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.ParseInt(l, 10, 32); err == nil {
				limit = int32(v)
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if v, err := strconv.ParseInt(o, 10, 32); err == nil {
				offset = int32(v)
			}
		}

		comments, err := commentRepo.ListByOwner(r.Context(), userID, limit, offset)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		responses, err := buildCommentResponses(r.Context(), comments, userID, userRepo, commentReactionRepo, reactionTypeRepo)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Success(w, responses)
	}
}

// @Summary Get user's comments by category
// @Description Fetch paginated comments of the authenticated user in a specific category
// @Tags comments
// @Security Bearer
// @Param category_id path int true "Category ID"
// @Param limit query int false "Limit" default(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} CommentResponse
// @Failure 401 {object} map[string]string
// @Router /user/comments/category/{category_id} [get]
func HandleGetUserCommentsByCategory(commentRepo *repository.CommentRepository, userRepo *repository.UserRepository, commentReactionRepo *repository.CommentReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		categoryIDStr := chi.URLParam(r, "category_id")
		categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid category_id")
			return
		}

		// Pagination
		limit, offset := int32(1000), int32(0)
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.ParseInt(l, 10, 32); err == nil {
				limit = int32(v)
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if v, err := strconv.ParseInt(o, 10, 32); err == nil {
				offset = int32(v)
			}
		}

		// Get comments by owner and category
		comments, err := commentRepo.ListByOwnerAndCategory(r.Context(), userID, categoryID, limit, offset)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		responses, err := buildCommentResponses(r.Context(), comments, userID, userRepo, commentReactionRepo, reactionTypeRepo)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Success(w, responses)
	}
}

// @Summary Get a comment by ID
// @Description Retrieve details of a specific comment from its ID
// @Tags comments
// @Security Bearer
// @Param comment_id path int true "Comment ID"
// @Success 200 {object} CommentResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /comments/{comment_id} [get]
func HandleGetComment(commentRepo *repository.CommentRepository, userRepo *repository.UserRepository, commentReactionRepo *repository.CommentReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		commentIDStr := chi.URLParam(r, "comment_id")
		commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid comment_id")
			return
		}

		comment, err := commentRepo.GetByID(r.Context(), commentID)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "comment not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}

		response, err := buildCommentResponse(r.Context(), comment, userID, userRepo, commentReactionRepo, reactionTypeRepo)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Success(w, response)
	}
}

// @Summary Create a comment on a post
// @Description Create a new comment for a specific post
// @Tags comments
// @Security Bearer
// @Param post_id path int true "Post ID"
// @Param request body CreateCommentRequest true "Comment data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /posts/{post_id}/comments [post]
func HandleCreateCommentOnPost(commentRepo *repository.CommentRepository, postRepo *repository.PostRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		postIDStr := chi.URLParam(r, "post_id")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid post_id")
			return
		}

		_, err = postRepo.GetByID(r.Context(), postID)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "post not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}

		var req CreateCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Text == nil || *req.Text == "" {
			ValidationError(w, "text is required")
			return
		}

		comment := &entity.Comment{
			PostID:  postID,
			OwnerID: userID,
			Text:    *req.Text,
			Image:   req.Image,
			Status:  false,
		}

		_, err = commentRepo.Create(r.Context(), comment)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Created(w, map[string]string{"message": "Comment created successfully!"})
	}
}

// HandleCreateReplyToComment creates a new reply to a comment.
// @Summary Reply to a comment
// @Description Create a reply to an existing comment
// @Tags comments
// @Security Bearer
// @Param comment_id path int true "Parent Comment ID"
// @Param request body CreateCommentRequest true "Reply data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /comments/{comment_id}/replies [post]
func HandleCreateReplyToComment(commentRepo *repository.CommentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		parentCommentIDStr := chi.URLParam(r, "comment_id")
		parentCommentID, err := strconv.ParseInt(parentCommentIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid comment_id")
			return
		}

		parentComment, err := commentRepo.GetByID(r.Context(), parentCommentID)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "comment not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}

		var req CreateCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Text == nil || *req.Text == "" {
			ValidationError(w, "text is required")
			return
		}

		comment := &entity.Comment{
			PostID:          parentComment.PostID,
			OwnerID:         userID,
			ParentCommentID: &parentCommentID,
			Text:            *req.Text,
			Image:           req.Image,
			Status:          false,
		}

		_, err = commentRepo.Create(r.Context(), comment)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Created(w, map[string]string{"message": "Reply created successfully!"})
	}
}

// HandleUpdateComment updates a comment or reply.
// @Summary Update a comment
// @Description Update the content of an existing comment or reply
// @Tags comments
// @Security Bearer
// @Param comment_id path int true "Comment ID"
// @Param request body UpdateCommentRequest true "Updated comment data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /comments/{comment_id} [put]
func HandleUpdateComment(commentRepo *repository.CommentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		commentIDStr := chi.URLParam(r, "comment_id")
		commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid comment_id")
			return
		}

		comment, err := commentRepo.GetByID(r.Context(), commentID)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "comment not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}

		if comment.OwnerID != userID {
			Forbidden(w, "you cannot update this comment")
			return
		}

		var req UpdateCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Text != nil && *req.Text != "" {
			comment.Text = *req.Text
		}
		if req.Image != nil {
			comment.Image = req.Image
		}

		err = commentRepo.Update(r.Context(), comment)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Success(w, map[string]string{"message": "Comment/Reply updated successfully!"})
	}
}

// HandleDeleteComment deletes a comment or reply.
// @Summary Delete a comment
// @Description Delete a comment and its associated reactions (owner only)
// @Tags comments
// @Security Bearer
// @Param comment_id path int true "Comment ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /comments/{comment_id} [delete]
func HandleDeleteComment(commentRepo *repository.CommentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		commentIDStr := chi.URLParam(r, "comment_id")
		commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid comment_id")
			return
		}

		comment, err := commentRepo.GetByID(r.Context(), commentID)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "comment not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}

		if comment.OwnerID != userID {
			Forbidden(w, "you cannot delete this comment")
			return
		}

		err = commentRepo.Delete(r.Context(), commentID)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Success(w, map[string]string{"message": "Comment/Reply deleted successfully!"})
	}
}

// @Summary React to a comment
// @Description Add or update new reaction to a comment or reply
// @Tags comments
// @Security Bearer
// @Param comment_id path int true "Comment ID"
// @Param request body ReactCommentRequest true "Reaction type"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /comments/{comment_id}/react [post]
func HandleReactToComment(commentRepo *repository.CommentRepository, commentReactionRepo *repository.CommentReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		commentIDStr := chi.URLParam(r, "comment_id")
		commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid comment_id")
			return
		}

		_, err = commentRepo.GetByID(r.Context(), commentID)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "comment not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}

		var req ReactCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.ReactionTypeID <= 0 {
			ValidationError(w, "reaction_type_id is required")
			return
		}

		_, err = reactionTypeRepo.GetByID(r.Context(), req.ReactionTypeID)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "reaction type not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}

		reaction := &entity.CommentReaction{
			CommentID:      commentID,
			OwnerID:        userID,
			ReactionTypeID: req.ReactionTypeID,
		}

		_, err = commentReactionRepo.Upsert(r.Context(), reaction)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		Success(w, map[string]string{"message": "Reaction recorded"})
	}
}

// buildCommentResponse builds a comment response from a comment entity with owner and reaction data
func buildCommentResponse(ctx context.Context, comment *entity.Comment, userID int64, userRepo *repository.UserRepository, commentReactionRepo *repository.CommentReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) (*CommentResponse, error) {
	owner, err := userRepo.GetByID(ctx, comment.OwnerID)
	if err != nil {
		return nil, err
	}

	totalReaction, err := commentReactionRepo.Count(ctx, comment.ID)
	if err != nil {
		return nil, err
	}

	var userReaction *ReactionInfo
	reaction, err := commentReactionRepo.GetByOwnerAndComment(ctx, userID, comment.ID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if reaction != nil {
		reactionType, err := reactionTypeRepo.GetByID(ctx, reaction.ReactionTypeID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if reactionType != nil {
			userReaction = &ReactionInfo{
				ReactionTypeID: reactionType.ID,
				Name:           reactionType.Name,
			}
		}
	}

	return &CommentResponse{
		CommentID:            comment.ID,
		CommentOwnerUsername: owner.Username,
		ProfilePicture:       owner.ProfilePicture,
		Text:                 comment.Text,
		Image:                comment.Image,
		CreatedAt:            comment.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:            comment.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		IsEdited:             comment.Status,
		TotalReaction:        totalReaction,
		UserReaction:         userReaction,
	}, nil
}

// buildCommentResponses builds multiple comment responses by calling buildCommentResponse for each comment
func buildCommentResponses(ctx context.Context, comments []*entity.Comment, userID int64, userRepo *repository.UserRepository, commentReactionRepo *repository.CommentReactionRepository, reactionTypeRepo *repository.ReactionTypeRepository) ([]*CommentResponse, error) {
	var responses []*CommentResponse
	for _, comment := range comments {
		response, err := buildCommentResponse(ctx, comment, userID, userRepo, commentReactionRepo, reactionTypeRepo)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}
