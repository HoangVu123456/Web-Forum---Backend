package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// ReactionRepository manages reactions on posts
type ReactionRepository struct {
	db *sql.DB
}

// NewReactionRepository creates a new ReactionRepository
func NewReactionRepository(db *sql.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

// Upsert sets a reaction for a post by owner, replacing existing one
func (r *ReactionRepository) Upsert(ctx context.Context, rec *entity.Reaction) (*entity.Reaction, error) {
	const q = `
        INSERT INTO reactions (post_id, owner_id, reaction_type_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (post_id, owner_id)
        DO UPDATE SET reaction_type_id = EXCLUDED.reaction_type_id
        RETURNING reaction_id
    `
	err := r.db.QueryRowContext(ctx, q, rec.PostID, rec.OwnerID, rec.ReactionTypeID).Scan(&rec.ID)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// GetByOwnerAndPost retrieves a reaction by owner and post IDs
func (r *ReactionRepository) GetByOwnerAndPost(ctx context.Context, ownerID, postID int64) (*entity.Reaction, error) {
	const q = `
        SELECT reaction_id, post_id, owner_id, reaction_type_id
        FROM reactions
        WHERE owner_id = $1 AND post_id = $2
    `
	row := r.db.QueryRowContext(ctx, q, ownerID, postID)
	return scanReaction(row)
}

// Delete removes a reaction by ID
func (r *ReactionRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM reactions WHERE reaction_id = $1`, id)
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

// CountByPost counts total reactions on a post
func (r *ReactionRepository) CountByPost(ctx context.Context, postID int64) (int64, error) {
	var count int64
	const q = `SELECT COUNT(*) FROM reactions WHERE post_id = $1`
	err := r.db.QueryRowContext(ctx, q, postID).Scan(&count)
	return count, err
}

// reactionRowScanner defines the interface for scanning reaction rows
type reactionRowScanner interface {
	Scan(dest ...any) error
}

// scanReaction scans a reaction from the given row scanner
func scanReaction(rs reactionRowScanner) (*entity.Reaction, error) {
	var rec entity.Reaction
	if err := rs.Scan(&rec.ID, &rec.PostID, &rec.OwnerID, &rec.ReactionTypeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &rec, nil
}
