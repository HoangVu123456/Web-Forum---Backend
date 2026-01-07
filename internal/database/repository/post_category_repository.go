package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// PostCategoryRepository links posts and categories.
type PostCategoryRepository struct {
	db *sql.DB
}

func NewPostCategoryRepository(db *sql.DB) *PostCategoryRepository {
	return &PostCategoryRepository{db: db}
}

func (r *PostCategoryRepository) Create(ctx context.Context, pc *entity.PostCategory) (*entity.PostCategory, error) {
	const q = `
        INSERT INTO post_categories (post_id, category_id)
        VALUES ($1, $2)
        ON CONFLICT (post_id, category_id) DO NOTHING
        RETURNING post_category_id
    `
	err := r.db.QueryRowContext(ctx, q, pc.PostID, pc.CategoryID).Scan(&pc.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pc, nil // already exists
		}
		return nil, err
	}
	return pc, nil
}

func (r *PostCategoryRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM post_categories WHERE post_category_id = $1`, id)
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

type postCategoryRowScanner interface {
	Scan(dest ...any) error
}

func scanPostCategory(rs postCategoryRowScanner) (*entity.PostCategory, error) {
	var pc entity.PostCategory
	if err := rs.Scan(&pc.ID, &pc.PostID, &pc.CategoryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &pc, nil
}
