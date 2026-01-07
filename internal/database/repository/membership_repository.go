package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// MembershipRepository manages user-category memberships.
type MembershipRepository struct {
	db *sql.DB
}

func NewMembershipRepository(db *sql.DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

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
			return m, nil // already exists
		}
		return nil, err
	}
	if joined.Valid {
		m.JoinedDate = joined.Time
	}
	return m, nil
}

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

type membershipRowScanner interface {
	Scan(dest ...any) error
}

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
