package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"my-chi-app/internal/database/repository"
	"my-chi-app/internal/domain/entity"

	"github.com/go-chi/chi/v5"
)

// UploadProfilePictureRequest is the payload for uploading a profile picture
type UploadProfilePictureRequest struct {
	ProfilePicture string `json:"profile_picture"`
}

// UpdateUsernameRequest is the payload for updating username
type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

// SubscribeRequest is the payload for subscribing to a category
type SubscribeRequest struct {
	Category string `json:"category"`
}

// UnsubscribeRequest is the payload for unsubscribing from a category
type UnsubscribeRequest struct {
	CategoryID int64 `json:"category_id"`
}

// MessageResponse is the payload response for simple messages
type MessageResponse struct {
	Message string `json:"message"`
}

// @Summary Update profile picture
// @Description Set or change user's profile picture
// @Tags users
// @Security Bearer
// @Param request body UploadProfilePictureRequest true "Profile picture URL"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /me/profile-picture [put]
func HandleUploadProfilePicture(userRepo *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		var req UploadProfilePictureRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.ProfilePicture == "" {
			ValidationError(w, "profile_picture is required")
			return
		}

		ctx := r.Context()

		// Get current user
		user, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "user not found")
				return
			}
			InternalError(w, "failed to fetch user")
			return
		}

		// Update profile picture
		if err := userRepo.UpdateProfilePicture(ctx, userID, req.ProfilePicture); err != nil {
			InternalError(w, "failed to update profile picture")
			return
		}

		user.ProfilePicture = &req.ProfilePicture

		Success(w, UserResponse{
			UserID:         user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Password:       user.Password,
			ProfilePicture: user.ProfilePicture,
			JoinedDate:     user.CreatedAt,
		})
	}
}

// @Summary Delete profile picture
// @Description Delete authenticated user's profile picture
// @Tags users
// @Security Bearer
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string
// @Router /me/profile-picture [delete]
func HandleDeleteProfilePicture(userRepo *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		ctx := r.Context()

		user, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "user not found")
				return
			}
			InternalError(w, "failed to fetch user")
			return
		}

		if err := userRepo.UpdateProfilePicture(ctx, userID, ""); err != nil {
			InternalError(w, "failed to delete profile picture")
			return
		}

		user.ProfilePicture = nil

		Success(w, UserResponse{
			UserID:         user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Password:       user.Password,
			ProfilePicture: user.ProfilePicture,
			JoinedDate:     user.CreatedAt,
		})
	}
}

// @Summary Update username
// @Description Change the authenticated user's username
// @Tags users
// @Security Bearer
// @Param request body UpdateUsernameRequest true "New username"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /me/username [put]
func HandleUpdateUsername(userRepo *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		var req UpdateUsernameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Username == "" {
			ValidationError(w, "username is required")
			return
		}

		ctx := r.Context()

		user, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "user not found")
				return
			}
			InternalError(w, "failed to fetch user")
			return
		}

		if err := userRepo.UpdateUsername(ctx, userID, req.Username); err != nil {
			if isDuplicateError(err) {
				Conflict(w, "username already exists")
				return
			}
			InternalError(w, "failed to update username")
			return
		}

		user.Username = req.Username

		Success(w, UserResponse{
			UserID:         user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Password:       user.Password,
			ProfilePicture: user.ProfilePicture,
			JoinedDate:     user.CreatedAt,
		})
	}
}

// @Summary Get user account
// @Description Retrieve public profile information for a specific user from user ID
// @Tags users
// @Security Bearer
// @Param user_id path int true "User ID"
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{user_id} [get]
func HandleGetAccount(userRepo *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		userIDStr := chi.URLParam(r, "user_id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid user_id")
			return
		}

		ctx := r.Context()

		user, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "user not found")
				return
			}
			InternalError(w, "failed to fetch user")
			return
		}

		Success(w, UserResponse{
			UserID:         user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Password:       user.Password,
			ProfilePicture: user.ProfilePicture,
			JoinedDate:     user.CreatedAt,
		})
	}
}

// @Summary Subscribe to category
// @Description Subscribe the authenticated user to a category
// @Tags users
// @Security Bearer
// @Param request body SubscribeRequest true "Category to subscribe"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /me/subscribe [post]
func HandleSubscribeCategory(userRepo *repository.UserRepository, categoryRepo *repository.CategoryRepository, membershipRepo *repository.MembershipRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		var req SubscribeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Category == "" {
			ValidationError(w, "category is required")
			return
		}

		ctx := r.Context()

		_, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "user not found")
				return
			}
			InternalError(w, "failed to fetch user")
			return
		}

		cat, err := categoryRepo.GetByName(ctx, req.Category)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "category not found")
				return
			}
			InternalError(w, "failed to fetch category")
			return
		}

		membership := &entity.Membership{
			CategoryID: cat.ID,
			UserID:     userID,
		}
		_, err = membershipRepo.Create(ctx, membership)
		if err != nil {
			InternalError(w, "failed to subscribe to category")
			return
		}

		Success(w, MessageResponse{
			Message: "Subscribe successfully!",
		})
	}
}

// @Summary Unsubscribe from category
// @Description Unsubscribe the authenticated user from a category
// @Tags users
// @Security Bearer
// @Param request body UnsubscribeRequest true "Category ID to unsubscribe"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /me/unsubscribe [post]
func HandleUnsubscribeCategory(membershipRepo *repository.MembershipRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		var req UnsubscribeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.CategoryID == 0 {
			ValidationError(w, "category_id is required")
			return
		}

		ctx := r.Context()

		_, err := membershipRepo.GetByUserAndCategory(ctx, userID, req.CategoryID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "membership not found")
				return
			}
			InternalError(w, "failed to fetch membership")
			return
		}

		if err := membershipRepo.DeleteByUserAndCategory(ctx, userID, req.CategoryID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "membership not found")
				return
			}
			InternalError(w, "failed to unsubscribe from category")
			return
		}

		Success(w, MessageResponse{
			Message: "Unsubscribe successfully!",
		})
	}
}

// HandleDeleteAccount deletes user account and all associated data.
// @Summary Delete account
// @Description Permanently delete the authenticated user's account and all associated data
// @Tags users
// @Security Bearer
// @Success 200 {object} MessageResponse
// @Failure 401 {object} map[string]string
// @Router /me [delete]
func HandleDeleteAccount(userRepo *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		ctx := r.Context()

		if err := userRepo.Delete(ctx, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "user not found")
				return
			}
			InternalError(w, "failed to delete account")
			return
		}

		Success(w, MessageResponse{
			Message: "Account delete successfully!",
		})
	}
}
