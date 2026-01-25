package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// CommentRepository manages comments and replies
type CommentRepository struct {
	db *sql.DB
}

// NewCommentRepository creates a new CommentRepository
func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create inserts a new comment into the database
func (r *CommentRepository) Create(ctx context.Context, c *entity.Comment) (*entity.Comment, error) {
	const q = `
        INSERT INTO comments (post_id, owner_id, parent_comment_id, text, image, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

// GetByID returns a comment by ID
func (r *CommentRepository) GetByID(ctx context.Context, id int64) (*entity.Comment, error) {
	const q = `
        SELECT comment_id, post_id, owner_id, parent_comment_id, text, image, created_at, updated_at, status
        FROM comments
        WHERE comment_id = $1
    `
	row := r.db.QueryRowContext(ctx, q, id)
	return scanComment(row)
}

// ListByPost returns comments for a specific post
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

// ListByParent returns replies to a specific comment
func (r *CommentRepository) ListByParent(ctx context.Context, parentID int64, limit, offset int32) ([]*entity.Comment, error) {
	const q = `
        SELECT comment_id, post_id, owner_id, parent_comment_id, text, image, created_at, updated_at, status
        FROM comments
        WHERE parent_comment_id = $1
        ORDER BY comment_id ASC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, q, parentID, limit, offset)
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

// ListByOwner returns all comments by a user
func (r *CommentRepository) ListByOwner(ctx context.Context, ownerID int64, limit, offset int32) ([]*entity.Comment, error) {
	const q = `
        SELECT comment_id, post_id, owner_id, parent_comment_id, text, image, created_at, updated_at, status
        FROM comments
        WHERE owner_id = $1
        ORDER BY comment_id DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, q, ownerID, limit, offset)
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

// ListByOwnerAndCategory returns comments by a user in a specific category
func (r *CommentRepository) ListByOwnerAndCategory(ctx context.Context, ownerID, categoryID int64, limit, offset int32) ([]*entity.Comment, error) {
	const q = `
        SELECT DISTINCT c.comment_id, c.post_id, c.owner_id, c.parent_comment_id, c.text, c.image, c.created_at, c.updated_at, c.status
        FROM comments c
        INNER JOIN posts p ON c.post_id = p.post_id
				WHERE c.owner_id = $1 AND p.category_id = $2
        ORDER BY c.comment_id DESC
        LIMIT $3 OFFSET $4
    `
	rows, err := r.db.QueryContext(ctx, q, ownerID, categoryID, limit, offset)
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

// Update modifies an existing comment
func (r *CommentRepository) Update(ctx context.Context, c *entity.Comment) error {
	const q = `
        UPDATE comments
				SET text = $2, image = $3, status = TRUE, updated_at = NOW()
        WHERE comment_id = $1
    `
	res, err := r.db.ExecContext(ctx, q, c.ID, c.Text, c.Image)
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

// Delete removes a comment by its ID
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

// commentRowScanner defines the interface for scanning comment rows
type commentRowScanner interface {
	Scan(dest ...any) error
}

// scanComment scans a comment from the given row scanner
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
