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

// ProductRepository menangani semua query ke tabel products.
type ProductRepository struct {
	db *pgxpool.Pool
}

// NewProductRepository membuat instance ProductRepository baru.
func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

// GetAll mengambil semua produk aktif (soft delete aware).
func (r *ProductRepository) GetAll(ctx context.Context) ([]model.Product, error) {
	query := `
		SELECT id, name, description, price, category, image_url, is_available, created_at, updated_at
		FROM   products
		WHERE  deleted_at IS NULL
		ORDER  BY name ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: get all products: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price,
			&p.Category, &p.ImageURL, &p.IsAvailable,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

// GetByID mengambil satu produk berdasarkan UUID.
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	query := `
		SELECT id, name, description, price, category, image_url, is_available, created_at, updated_at
		FROM   products
		WHERE  id = $1 AND deleted_at IS NULL
	`
	var p model.Product
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price,
		&p.Category, &p.ImageURL, &p.IsAvailable,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("product not found: %s", id)
		}
		return nil, fmt.Errorf("repository: get product by id: %w", err)
	}
	return &p, nil
}

// Create menyimpan produk baru ke DB.
func (r *ProductRepository) Create(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error) {
	query := `
		INSERT INTO products (id, name, description, price, category, image_url)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		RETURNING id, name, description, price, category, image_url, is_available, created_at, updated_at
	`
	var p model.Product
	err := r.db.QueryRow(ctx, query,
		req.Name, req.Description, req.Price, req.Category, req.ImageURL,
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price,
		&p.Category, &p.ImageURL, &p.IsAvailable,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create product: %w", err)
	}
	return &p, nil
}

// SoftDelete menandai produk sebagai deleted tanpa menghapus dari DB.
func (r *ProductRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: soft delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found or already deleted: %s", id)
	}
	return nil
}
