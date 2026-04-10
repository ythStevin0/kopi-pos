package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ythStevin0/kopi-pos/services/api/internal/model"
)

// IngredientRepository menangani semua query ke tabel ingredients.
type IngredientRepository struct {
	db *pgxpool.Pool
}

// NewIngredientRepository membuat instance IngredientRepository baru.
func NewIngredientRepository(db *pgxpool.Pool) *IngredientRepository {
	return &IngredientRepository{db: db}
}

// GetAll mengambil semua ingredient aktif.
func (r *IngredientRepository) GetAll(ctx context.Context) ([]model.Ingredient, error) {
	query := `
		SELECT id, name, unit, stock, min_stock, created_at, updated_at
		FROM   ingredients
		WHERE  deleted_at IS NULL
		ORDER  BY name ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: get all ingredients: %w", err)
	}
	defer rows.Close()

	var ingredients []model.Ingredient
	for rows.Next() {
		var i model.Ingredient
		if err := rows.Scan(
			&i.ID, &i.Name, &i.Unit, &i.Stock, &i.MinStock,
			&i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan ingredient: %w", err)
		}
		ingredients = append(ingredients, i)
	}
	return ingredients, rows.Err()
}

// GetByID mengambil satu ingredient berdasarkan UUID.
func (r *IngredientRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Ingredient, error) {
	query := `
		SELECT id, name, unit, stock, min_stock, created_at, updated_at
		FROM   ingredients
		WHERE  id = $1 AND deleted_at IS NULL
	`
	var i model.Ingredient
	err := r.db.QueryRow(ctx, query, id).Scan(
		&i.ID, &i.Name, &i.Unit, &i.Stock, &i.MinStock,
		&i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ingredient not found: %s", id)
		}
		return nil, fmt.Errorf("repository: get ingredient by id: %w", err)
	}
	return &i, nil
}

// SoftDelete menandai ingredient sebagai deleted.
func (r *IngredientRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE ingredients SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: soft delete ingredient: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ingredient not found or already deleted: %s", id)
	}
	return nil
}
