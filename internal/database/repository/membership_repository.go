package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// MembershipRepository manages user-category memberships
type MembershipRepository struct {
	db *sql.DB
}

// NewMembershipRepository creates a new MembershipRepository
func NewMembershipRepository(db *sql.DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

// Create adds a new membership for a user in a category
func (r *MembershipRepository) Create(ctx context.Context, m *entity.Membership) (*entity.Membership, error) {
	const q = `
        INSERT INTO memberships (category_id, user_id)
        VALUES ($1, $2)
        ON CONFLICT (category_id, user_id) DO NOTHING
        RETURNING membership_id, joined_date
    `
	var joined sql.NullTime
	err := r.db.QueryRowContext(ctx, q, m.CategoryID, m.UserID).Scan(&m.ID, &joined)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m, nil
		}
		return nil, err
	}
	if joined.Valid {
		m.JoinedDate = joined.Time
	}
	return m, nil
}

// Delete removes a membership by its ID
func (r *MembershipRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM memberships WHERE membership_id = $1`, id)
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

// GetByUserAndCategory retrieves a membership by user ID and category ID
func (r *MembershipRepository) GetByUserAndCategory(ctx context.Context, userID, categoryID int64) (*entity.Membership, error) {
	const q = `
				SELECT membership_id, category_id, user_id, joined_date
				FROM memberships
				WHERE user_id = $1 AND category_id = $2
			`
	row := r.db.QueryRowContext(ctx, q, userID, categoryID)
	return scanMembership(row)
}

// DeleteByUserAndCategory removes membership by user and category IDs
func (r *MembershipRepository) DeleteByUserAndCategory(ctx context.Context, userID, categoryID int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM memberships WHERE user_id = $1 AND category_id = $2`, userID, categoryID)
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

// GetByUserID returns all memberships for a user
func (r *MembershipRepository) GetByUserID(ctx context.Context, userID int64) ([]*entity.Membership, error) {
	const q = `
        SELECT membership_id, category_id, user_id, joined_date
        FROM memberships
        WHERE user_id = $1
        ORDER BY joined_date DESC
    `
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []*entity.Membership
	for rows.Next() {
		m, err := scanMembership(rows)
		if err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return memberships, nil
}

// membershipRowScanner defines the interface for scanning membership rows
type membershipRowScanner interface {
	Scan(dest ...any) error
}

// scanMembership scans a membership from the given row scanner
func scanMembership(rs membershipRowScanner) (*entity.Membership, error) {
	var m entity.Membership
	if err := rs.Scan(&m.ID, &m.CategoryID, &m.UserID, &m.JoinedDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &m, nil
}
