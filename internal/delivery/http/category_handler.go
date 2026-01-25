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

// CreateCategoryRequest is the payload for creating a new category.
type CreateCategoryRequest struct {
	Category string `json:"category"`
}

// CategoryResponse is the payload response when returning category information
type CategoryResponse struct {
	CategoryID int64  `json:"category_id"`
	Category   string `json:"category"`
}

// CategoryCreatedResponse is the response after successfully creating a category.
type CategoryCreatedResponse struct {
	Message string `json:"message"`
}

// Swagger annotations:
// @Summary Get all categories
// @Description Retrieve a list containing all forum categories
// @Tags categories
// @Security Bearer
// @Success 200 {array} CategoryResponse
// @Failure 401 {object} map[string]string
// @Router /categories [get]
func HandleGetAllCategories(categoryRepo *repository.CategoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		ctx := r.Context()

		categories, err := categoryRepo.List(ctx)
		if err != nil {
			InternalError(w, "failed to fetch categories")
			return
		}

		response := make([]CategoryResponse, len(categories))
		for i, cat := range categories {
			response[i] = CategoryResponse{
				CategoryID: cat.ID,
				Category:   cat.Category,
			}
		}

		Success(w, response)
	}
}

// Swagger annotations:
// @Summary Get user's subscribed categories
// @Description Retrieve a list of categories the authenticated user is subscribed to
// @Tags categories
// @Security Bearer
// @Success 200 {array} CategoryResponse
// @Failure 401 {object} map[string]string
// @Router /user/categories [get]
func HandleGetUserCategories(membershipRepo *repository.MembershipRepository, categoryRepo *repository.CategoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		ctx := r.Context()

		memberships, err := membershipRepo.GetByUserID(ctx, userID)
		if err != nil {
			InternalError(w, "failed to fetch subscriptions")
			return
		}

		response := make([]CategoryResponse, 0)
		for _, membership := range memberships {
			cat, err := categoryRepo.GetByID(ctx, membership.CategoryID)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					InternalError(w, "failed to fetch category")
					return
				}
				continue
			}
			response = append(response, CategoryResponse{
				CategoryID: cat.ID,
				Category:   cat.Category,
			})
		}

		Success(w, response)
	}
}

// Swagger annotations:
// @Summary Get a category by ID
// @Description Retrieve details of a specific category from its ID
// @Tags categories
// @Security Bearer
// @Param category_id path int true "Category ID"
// @Success 200 {object} CategoryResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /categories/{category_id} [get]
func HandleGetCategoryByID(categoryRepo *repository.CategoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserID(r.Context())
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

		category, err := categoryRepo.GetByID(ctx, categoryID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				NotFound(w, "category not found")
				return
			}
			InternalError(w, "failed to fetch category")
			return
		}

		Success(w, CategoryResponse{
			CategoryID: category.ID,
			Category:   category.Category,
		})
	}
}

// Swagger annotations:
// @Summary Create a new category
// @Description Create a new category for the forum
// @Tags categories
// @Security Bearer
// @Param request body CreateCategoryRequest true "Category data"
// @Success 201 {object} CategoryCreatedResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /categories [post]
func HandleCreateCategory(categoryRepo *repository.CategoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "user not authenticated")
			return
		}

		var req CreateCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Category == "" {
			ValidationError(w, "category is required")
			return
		}

		ctx := r.Context()

		category := &entity.Category{Category: req.Category}
		_, err := categoryRepo.Create(ctx, category)
		if err != nil {
			if isDuplicateError(err) {
				Conflict(w, "category already exists")
				return
			}
			InternalError(w, "failed to create category")
			return
		}

		Success(w, CategoryCreatedResponse{
			Message: "Category created successfully!",
		})
	}
}
