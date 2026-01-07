package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// PostRepository manages posts.
type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, p *entity.Post) (*entity.Post, error) {
	const q = `
        INSERT INTO posts (owner_id, headline, text, image, status)
        VALUES ($1, $2, $3, $4, $5)
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

	err := r.db.QueryRowContext(ctx, q, p.OwnerID, p.Headline, text, image, p.Status).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PostRepository) GetByID(ctx context.Context, id int64) (*entity.Post, error) {
	const q = `
        SELECT post_id, owner_id, headline, text, image, created_at, updated_at, status
        FROM posts
        WHERE post_id = $1
    `
	row := r.db.QueryRowContext(ctx, q, id)
	return scanPost(row)
}

func (r *PostRepository) List(ctx context.Context, limit, offset int32) ([]*entity.Post, error) {
	const q = `
        SELECT post_id, owner_id, headline, text, image, created_at, updated_at, status
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

type postRowScanner interface {
	Scan(dest ...any) error
}

func scanPost(rs postRowScanner) (*entity.Post, error) {
	var (
		p     entity.Post
		text  sql.NullString
		image sql.NullString
	)

	if err := rs.Scan(&p.ID, &p.OwnerID, &p.Headline, &text, &image, &p.CreatedAt, &p.UpdatedAt, &p.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	if text.Valid {
		p.Text = &text.String
	}
	if image.Valid {
		p.Image = &image.String
	}
	return &p, nil
}
