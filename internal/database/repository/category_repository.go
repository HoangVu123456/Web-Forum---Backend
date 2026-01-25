package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// CategoryRepository manages categories.
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create inserts a new category into the database
func (r *CategoryRepository) Create(ctx context.Context, c *entity.Category) (*entity.Category, error) {
	const q = `
        INSERT INTO categories (category)
        VALUES ($1)
        RETURNING category_id
    `

	err := r.db.QueryRowContext(ctx, q, c.Category).Scan(&c.ID)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetByID returns a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*entity.Category, error) {
	const q = `
        SELECT category_id, category
        FROM categories
        WHERE category_id = $1
    `
	row := r.db.QueryRowContext(ctx, q, id)
	return scanCategory(row)
}

// List returns all categories
func (r *CategoryRepository) List(ctx context.Context) ([]*entity.Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT category_id, category FROM categories ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Category
	for rows.Next() {
		c, err := scanCategory(rows)
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

// GetByName returns a category by name.
func (r *CategoryRepository) GetByName(ctx context.Context, name string) (*entity.Category, error) {
	const q = `
        SELECT category_id, category
        FROM categories
        WHERE category = $1
    `
	row := r.db.QueryRowContext(ctx, q, name)
	return scanCategory(row)
}

// categoryRowScanner defines the interface for scanning category rows
type categoryRowScanner interface {
	Scan(dest ...any) error
}

// scanCategory scans a category from the given row scanner
func scanCategory(rs categoryRowScanner) (*entity.Category, error) {
	var c entity.Category
	if err := rs.Scan(&c.ID, &c.Category); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &c, nil
}
