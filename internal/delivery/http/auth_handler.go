package http

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"my-chi-app/internal/database/repository"
	"my-chi-app/internal/domain/entity"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest is the payload request for registering a new user
type RegisterRequest struct {
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Password   string     `json:"password"`
	JoinedDate *time.Time `json:"joined_date,omitempty"`
}

// LoginRequest is the payload request for logging in an existing user
type LoginRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password"`
}

// UserResponse is the response returned if registration or login is successful
type UserResponse struct {
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	ProfilePicture *string   `json:"profile_picture,omitempty"`
	JoinedDate     time.Time `json:"joined_date"`
	Token          string    `json:"token"`
}

// LogoutResponse is the payload response for logging out a user
type LogoutResponse struct {
	Message string `json:"message"`
}

// Swagger annotations:
// @Summary Register a new user
// @Description Create a new user account with username, email, and password and return the user details with JWT token
// @Tags auth
// @Param request body RegisterRequest true "Registration data"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func HandleRegister(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Username == "" || req.Email == "" || req.Password == "" {
			ValidationError(w, "username, email, and password are required")
			return
		}
		if req.Email != "" && !isValidEmail(req.Email) {
			ValidationError(w, "invalid email format")
			return
		}
		if len(req.Password) < 8 {
			ValidationError(w, "password must be at least 8 characters")
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			InternalError(w, "failed to hash password")
			return
		}

		user := &entity.User{
			Username: req.Username,
			Email:    req.Email,
			Password: string(hashedPassword),
		}
		ctx := r.Context()

		user, err = userRepo.Create(ctx, user)
		if err != nil {
			if isDuplicateError(err) {
				Conflict(w, "email or username already exists")
				return
			}
			InternalError(w, "failed to create user")
			return
		}

		token, _, err := createToken(ctx, tokenRepo, user.ID, jwtSecret)
		if err != nil {
			InternalError(w, "failed to create token")
			return
		}

		Success(w, UserResponse{
			UserID:         user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Password:       user.Password,
			ProfilePicture: user.ProfilePicture,
			JoinedDate:     user.CreatedAt,
			Token:          token,
		})
	}
}

// Swagger annotations:
// @Summary User login
// @Description Authenticate user with email/username and password then return user details with JWT token
// @Tags auth
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func HandleLogin(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "invalid request body")
			return
		}

		if req.Password == "" {
			ValidationError(w, "password is required")
			return
		}
		if req.Email == "" && req.Username == "" {
			ValidationError(w, "email or username is required")
			return
		}

		ctx := r.Context()
		var user *entity.User
		var err error

		if req.Email != "" {
			user, err = userRepo.GetByEmail(ctx, req.Email)
		} else {
			user, err = userRepo.GetByUsername(ctx, req.Username)
		}

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				Unauthorized(w, "invalid credentials")
				return
			}
			InternalError(w, "failed to fetch user")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			Unauthorized(w, "invalid credentials")
			return
		}

		token, _, err := createToken(ctx, tokenRepo, user.ID, jwtSecret)
		if err != nil {
			InternalError(w, "failed to create token")
			return
		}

		Success(w, UserResponse{
			UserID:         user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Password:       user.Password,
			ProfilePicture: user.ProfilePicture,
			JoinedDate:     user.CreatedAt,
			Token:          token,
		})
	}
}

// Swagger annotations:
// @Summary User logout
// @Description Logout of the current session and invalidate the JWT token
// @Tags auth
// @Success 200 {object} LogoutResponse
// @Failure 401 {object} map[string]string
// @Router /auth/logout [post]
func HandleLogOut(tokenRepo *repository.TokenRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract JWT token
		token := extractToken(r)
		if token == "" {
			Unauthorized(w, "missing authorization token")
			return
		}

		ctx := r.Context()

		// Find and delete token in database
		t, err := tokenRepo.GetByToken(ctx, token)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				Unauthorized(w, "invalid token")
				return
			}
			InternalError(w, "failed to fetch token")
			return
		}

		if err := tokenRepo.DeleteByID(ctx, t.ID); err != nil {
			InternalError(w, "failed to delete token")
			return
		}

		Success(w, LogoutResponse{
			Message: "Logout successfully!",
		})
	}
}

// Swagger annotations:
// @Summary Verify authentication status
// @Description Check if the current authentication token is valid and return user ID
// @Tags auth
// @Security Bearer
// @Success 200 {object} VerifyResponse
// @Failure 401 {object} map[string]string
// @Router /auth/verify [get]
func HandleVerifyAuth(userRepo *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "invalid or missing user ID")
			return
		}

		response := VerifyResponse{
			UserID: userID,
			Valid:  true,
			Status: "authenticated",
		}
		JSONResponse(w, http.StatusOK, response)
	}
}

// createToken generates a JWT token and stores it in the database
// Token specifies userID as the subject and expires in 24 hours
func createToken(ctx context.Context, tokenRepo *repository.TokenRepository, userID int64, jwtSecret string) (string, time.Time, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	jti := hex.EncodeToString(b)

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        jti,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	t := &entity.Token{
		UserID:    userID,
		Token:     signed,
		ExpiresAt: expiresAt,
	}

	if _, err := tokenRepo.Create(ctx, t); err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

// extractToken gets bearer token from Authorization header.
func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

// isDuplicateError checks for a unique constraint violation error
// Validation code: 23505 in PostgreSQL
func isDuplicateError(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate") ||
		contains(err.Error(), "unique") ||
		contains(err.Error(), "23505"))
}

// isValidEmail performs a format check for email input
func isValidEmail(email string) bool {
	at := false
	dot := false
	for i := 0; i < len(email); i++ {
		if email[i] == '@' {
			at = true
		}
		if at && email[i] == '.' {
			dot = true
		}
	}
	return at && dot
}

// contains checks if substr is in s with 3 possible positions
// If doesn't match start or end, it calls findSubstring for middle check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

// findSubstring checks if substr is in s (general case)
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// VerifyResponse is the response returned when verifying auth status
type VerifyResponse struct {
	UserID int64  `json:"user_id"`
	Valid  bool   `json:"valid"`
	Status string `json:"status"`
}
