package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// UserRepository provides CRUD operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and returns the created user data
func (r *UserRepository) Create(ctx context.Context, u *entity.User) (*entity.User, error) {
	const q = `
        INSERT INTO users (username, email, password, profile_picture)
        VALUES ($1, $2, $3, $4)
        RETURNING user_id, created_at
    `

	var profile *string
	if u.ProfilePicture != nil && *u.ProfilePicture != "" {
		profile = u.ProfilePicture
	}

	err := r.db.QueryRowContext(ctx, q, u.Username, u.Email, u.Password, profile).
		Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// GetByID returns a user by primary key
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	const q = `
        SELECT user_id, username, email, password, profile_picture, created_at
        FROM users
        WHERE user_id = $1
    `

	row := r.db.QueryRowContext(ctx, q, id)
	return scanUser(row)
}

// GetByEmail returns a user matching the email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	const q = `
        SELECT user_id, username, email, password, profile_picture, created_at
        FROM users
        WHERE email = $1
    `

	row := r.db.QueryRowContext(ctx, q, email)
	return scanUser(row)
}

// GetByUsername returns a user matching the username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	const q = `
        SELECT user_id, username, email, password, profile_picture, created_at
        FROM users
        WHERE username = $1
    `

	row := r.db.QueryRowContext(ctx, q, username)
	return scanUser(row)
}

// List returns users ordered by newest first with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int32) ([]*entity.User, error) {
	const q = `
        SELECT user_id, username, email, password, profile_picture, created_at
        FROM users
        ORDER BY user_id DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM users WHERE user_id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateProfilePicture updates user's profile picture
func (r *UserRepository) UpdateProfilePicture(ctx context.Context, userID int64, picture string) error {
	const q = `UPDATE users SET profile_picture = NULLIF($2, '') WHERE user_id = $1`
	res, err := r.db.ExecContext(ctx, q, userID, picture)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateUsername updates user's username
func (r *UserRepository) UpdateUsername(ctx context.Context, userID int64, username string) error {
	const q = `UPDATE users SET username = $2 WHERE user_id = $1`
	res, err := r.db.ExecContext(ctx, q, userID, username)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// rowScanner defines the interface for scanning user rows
type rowScanner interface {
	Scan(dest ...any) error
}

// scanUser scans a user from the given row scanner
func scanUser(rs rowScanner) (*entity.User, error) {
	var (
		u       entity.User
		profile sql.NullString
	)

	if err := rs.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &profile, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	if profile.Valid {
		u.ProfilePicture = &profile.String
	}

	return &u, nil
}
