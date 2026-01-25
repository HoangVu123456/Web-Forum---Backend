package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// ReactionTypeRepository manages reaction types
type ReactionTypeRepository struct {
	db *sql.DB
}

// NewReactionTypeRepository creates a new ReactionTypeRepository
func NewReactionTypeRepository(db *sql.DB) *ReactionTypeRepository {
	return &ReactionTypeRepository{db: db}
}

// Create inserts a new reaction type into the database
func (r *ReactionTypeRepository) Create(ctx context.Context, rt *entity.ReactionType) (*entity.ReactionType, error) {
	const q = `
        INSERT INTO reaction_types (name, image)
        VALUES ($1, $2)
        RETURNING reaction_type_id
    `

	var image sql.NullString
	if rt.Image != nil && *rt.Image != "" {
		image.String, image.Valid = *rt.Image, true
	}

	err := r.db.QueryRowContext(ctx, q, rt.Name, image).Scan(&rt.ID)
	if err != nil {
		return nil, err
	}
	return rt, nil
}

// GetByID returns a reaction type by ID
func (r *ReactionTypeRepository) GetByID(ctx context.Context, id int64) (*entity.ReactionType, error) {
	const q = `
        SELECT reaction_type_id, name, image
        FROM reaction_types
        WHERE reaction_type_id = $1
    `
	row := r.db.QueryRowContext(ctx, q, id)
	return scanReactionType(row)
}

// List returns all reaction types
func (r *ReactionTypeRepository) List(ctx context.Context) ([]*entity.ReactionType, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT reaction_type_id, name, image FROM reaction_types ORDER BY reaction_type_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.ReactionType
	for rows.Next() {
		rt, err := scanReactionType(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, rt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// reactionTypeRowScanner defines the interface for scanning reaction type rows
type reactionTypeRowScanner interface {
	Scan(dest ...any) error
}

// scanReactionType scans a reaction type from the given row scanner
func scanReactionType(rs reactionTypeRowScanner) (*entity.ReactionType, error) {
	var (
		rt    entity.ReactionType
		image sql.NullString
	)

	if err := rs.Scan(&rt.ID, &rt.Name, &image); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	if image.Valid {
		rt.Image = &image.String
	}
	return &rt, nil
}
