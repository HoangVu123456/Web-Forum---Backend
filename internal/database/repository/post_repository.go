package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// PostRepository manages posts
type PostRepository struct {
	db *sql.DB
}

// NewPostRepository creates a new PostRepository
func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// Create inserts a new post into the database
func (r *PostRepository) Create(ctx context.Context, p *entity.Post) (*entity.Post, error) {
	const q = `
				INSERT INTO posts (owner_id, category_id, headline, text, image, status)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING post_id, created_at, updated_at
    `

	var text sql.NullString
	if p.Text != nil && *p.Text != "" {
		text.String, text.Valid = *p.Text, true
	}
	var image sql.NullString
	if p.Image != nil && *p.Image != "" {
		image.String, image.Valid = *p.Image, true
	}

	err := r.db.QueryRowContext(ctx, q, p.OwnerID, p.CategoryID, p.Headline, text, image, p.Status).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PostRepository) GetByID(ctx context.Context, id int64) (*entity.Post, error) {
	const q = `
				SELECT post_id, owner_id, category_id, headline, text, image, created_at, updated_at, status
        FROM posts
        WHERE post_id = $1
    `
	row := r.db.QueryRowContext(ctx, q, id)
	return scanPost(row)
}

// List returns all posts with pagination
func (r *PostRepository) List(ctx context.Context, limit, offset int32) ([]*entity.Post, error) {
	const q = `
				SELECT post_id, owner_id, category_id, headline, text, image, created_at, updated_at, status
        FROM posts
        ORDER BY post_id DESC
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// Delete removes a post by ID
func (r *PostRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE post_id = $1`, id)
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

// GetByOwner returns posts created by a user
func (r *PostRepository) GetByOwner(ctx context.Context, ownerID int64, limit, offset int32) ([]*entity.Post, error) {
	const q = `
				SELECT post_id, owner_id, category_id, headline, text, image, created_at, updated_at, status
        FROM posts
        WHERE owner_id = $1
        ORDER BY post_id DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// GetByCategory returns posts in a category
func (r *PostRepository) GetByCategory(ctx context.Context, categoryID int64, limit, offset int32) ([]*entity.Post, error) {
	const q = `
				SELECT p.post_id, p.owner_id, p.category_id, p.headline, p.text, p.image, p.created_at, p.updated_at, p.status
				FROM posts p
				WHERE p.category_id = $1
				ORDER BY p.post_id DESC
				LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, q, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// GetByOwnerAndCategory returns user's posts in a specific category
func (r *PostRepository) GetByOwnerAndCategory(ctx context.Context, ownerID, categoryID int64, limit, offset int32) ([]*entity.Post, error) {
	const q = `
				SELECT p.post_id, p.owner_id, p.category_id, p.headline, p.text, p.image, p.created_at, p.updated_at, p.status
				FROM posts p
				WHERE p.owner_id = $1 AND p.category_id = $2
				ORDER BY p.post_id DESC
				LIMIT $3 OFFSET $4
    `
	rows, err := r.db.QueryContext(ctx, q, ownerID, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// Update modifies an existing post
func (r *PostRepository) Update(ctx context.Context, p *entity.Post) error {
	const q = `
        UPDATE posts
				SET headline = $2, text = $3, image = $4, status = TRUE, updated_at = NOW()
        WHERE post_id = $1
    `
	res, err := r.db.ExecContext(ctx, q, p.ID, p.Headline, p.Text, p.Image)
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

// postRowScanner defines the interface for scanning post rows
type postRowScanner interface {
	Scan(dest ...any) error
}

// scanPost scans a post from the given row scanner
func scanPost(rs postRowScanner) (*entity.Post, error) {
	var (
		p          entity.Post
		categoryID int64
		text       sql.NullString
		image      sql.NullString
	)

	if err := rs.Scan(&p.ID, &p.OwnerID, &categoryID, &p.Headline, &text, &image, &p.CreatedAt, &p.UpdatedAt, &p.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	p.CategoryID = categoryID
	if text.Valid {
		p.Text = &text.String
	}
	if image.Valid {
		p.Image = &image.String
	}
	return &p, nil
}
