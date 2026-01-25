package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// CommentReactionRepository manages reactions on comments and replies
type CommentReactionRepository struct {
	db *sql.DB
}

// NewCommentReactionRepository creates a new CommentReactionRepository
func NewCommentReactionRepository(db *sql.DB) *CommentReactionRepository {
	return &CommentReactionRepository{db: db}
}

// Upsert sets a reaction for a comment by user, updating it if it already exists
func (r *CommentReactionRepository) Upsert(ctx context.Context, rec *entity.CommentReaction) (*entity.CommentReaction, error) {
	const q = `
        INSERT INTO comment_reactions (comment_id, owner_id, reaction_type_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (comment_id, owner_id)
        DO UPDATE SET reaction_type_id = EXCLUDED.reaction_type_id
        RETURNING comment_reaction_id
    `
	err := r.db.QueryRowContext(ctx, q, rec.CommentID, rec.OwnerID, rec.ReactionTypeID).Scan(&rec.ID)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// GetByOwnerAndComment retrieves a reaction by user and comment IDs
func (r *CommentReactionRepository) GetByOwnerAndComment(ctx context.Context, ownerID, commentID int64) (*entity.CommentReaction, error) {
	const q = `
        SELECT comment_reaction_id, comment_id, owner_id, reaction_type_id
        FROM comment_reactions
        WHERE owner_id = $1 AND comment_id = $2
    `
	row := r.db.QueryRowContext(ctx, q, ownerID, commentID)
	return scanCommentReaction(row)
}

// Count returns the total number of reactions on a comment
func (r *CommentReactionRepository) Count(ctx context.Context, commentID int64) (int64, error) {
	const q = `
        SELECT COUNT(comment_reaction_id)
        FROM comment_reactions
        WHERE comment_id = $1
    `
	var count int64
	err := r.db.QueryRowContext(ctx, q, commentID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return count, nil
}

// Delete removes a reaction by its ID
func (r *CommentReactionRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM comment_reactions WHERE comment_reaction_id = $1`, id)
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

// commentReactionRowScanner defines the interface for scanning comment reaction rows
type commentReactionRowScanner interface {
	Scan(dest ...any) error
}

// scanCommentReaction scans a comment reaction from the given row scanner
func scanCommentReaction(rs commentReactionRowScanner) (*entity.CommentReaction, error) {
	var rec entity.CommentReaction
	if err := rs.Scan(&rec.ID, &rec.CommentID, &rec.OwnerID, &rec.ReactionTypeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &rec, nil
}
