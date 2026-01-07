package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// CommentRepository manages comments.
type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, c *entity.Comment) (*entity.Comment, error) {
	const q = `
        INSERT INTO comments (post_id, owner_id, parent_comment_id, text, image, status)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING comment_id, created_at, updated_at
    `

	var parent sql.NullInt64
	if c.ParentCommentID != nil {
		parent.Int64, parent.Valid = *c.ParentCommentID, true
	}
	var image sql.NullString
	if c.Image != nil && *c.Image != "" {
		image.String, image.Valid = *c.Image, true
	}

	err := r.db.QueryRowContext(ctx, q, c.PostID, c.OwnerID, parent, c.Text, image, c.Status).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id int64) (*entity.Comment, error) {
	const q = `
        SELECT comment_id, post_id, owner_id, parent_comment_id, text, image, created_at, updated_at, status
        FROM comments
        WHERE comment_id = $1
    `
	row := r.db.QueryRowContext(ctx, q, id)
	return scanComment(row)
}

func (r *CommentRepository) ListByPost(ctx context.Context, postID int64, limit, offset int32) ([]*entity.Comment, error) {
	const q = `
        SELECT comment_id, post_id, owner_id, parent_comment_id, text, image, created_at, updated_at, status
        FROM comments
        WHERE post_id = $1
        ORDER BY comment_id ASC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, q, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Comment
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *CommentRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE comment_id = $1`, id)
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

type commentRowScanner interface {
	Scan(dest ...any) error
}

func scanComment(rs commentRowScanner) (*entity.Comment, error) {
	var (
		c      entity.Comment
		parent sql.NullInt64
		image  sql.NullString
	)

	if err := rs.Scan(&c.ID, &c.PostID, &c.OwnerID, &parent, &c.Text, &image, &c.CreatedAt, &c.UpdatedAt, &c.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	if parent.Valid {
		c.ParentCommentID = &parent.Int64
	}
	if image.Valid {
		c.Image = &image.String
	}
	return &c, nil
}
