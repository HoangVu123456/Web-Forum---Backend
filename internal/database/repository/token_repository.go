package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"my-chi-app/internal/domain/entity"
)

// TokenRepository manages auth tokens
type TokenRepository struct {
	db *sql.DB
}

// NewTokenRepository creates a new TokenRepository
func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// Create inserts a new token into the database
func (r *TokenRepository) Create(ctx context.Context, t *entity.Token) (*entity.Token, error) {
	const q = `
        INSERT INTO tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
        RETURNING token_id, expires_at
    `

	err := r.db.QueryRowContext(ctx, q, t.UserID, t.Token, t.ExpiresAt).
		Scan(&t.ID, &t.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// GetByToken retrieves a token by its string value
func (r *TokenRepository) GetByToken(ctx context.Context, token string) (*entity.Token, error) {
	const q = `
        SELECT token_id, user_id, token, expires_at
        FROM tokens
        WHERE token = $1
    `
	row := r.db.QueryRowContext(ctx, q, token)
	return scanToken(row)
}

// DeleteByID removes a token by its ID
func (r *TokenRepository) DeleteByID(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tokens WHERE token_id = $1`, id)
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

// PurgeExpired deletes all tokens that have expired before the cutoff time
func (r *TokenRepository) PurgeExpired(ctx context.Context, cutoff time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tokens WHERE expires_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// tokenRowScanner defines the interface for scanning token rows
type tokenRowScanner interface {
	Scan(dest ...any) error
}

// scanToken scans a token from the given row scanner
func scanToken(rs tokenRowScanner) (*entity.Token, error) {
	var t entity.Token
	if err := rs.Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &t, nil
}
