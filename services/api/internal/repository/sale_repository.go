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

// SaleRepository menangani semua query ke tabel sales & sale_items.
type SaleRepository struct {
	db *pgxpool.Pool
}

// NewSaleRepository membuat instance SaleRepository baru.
func NewSaleRepository(db *pgxpool.Pool) *SaleRepository {
	return &SaleRepository{db: db}
}

// IdempotencyKeyExists mengecek apakah transaksi dengan key ini sudah pernah diproses.
func (r *SaleRepository) IdempotencyKeyExists(ctx context.Context, key uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM sales WHERE idempotency_key = $1)`
	err := r.db.QueryRow(ctx, query, key).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("repository: check idempotency: %w", err)
	}
	return exists, nil
}

// GetRecipeByProductID mengambil semua bahan yang dibutuhkan untuk satu produk.
// Dijalankan di dalam transaksi aktif.
func (r *SaleRepository) GetRecipeByProductID(
	ctx context.Context,
	tx pgx.Tx,
	productID uuid.UUID,
	qty int,
) ([]model.RecipeDetail, error) {
	query := `
		SELECT
			r.ingredient_id,
			i.name,
			i.unit,
			r.usage_quantity,
			r.usage_quantity * $2 AS total_deduct
		FROM   recipes r
		JOIN   ingredients i ON i.id = r.ingredient_id
		WHERE  r.product_id = $1
		  AND  i.deleted_at IS NULL
	`
	rows, err := tx.Query(ctx, query, productID, qty)
	if err != nil {
		return nil, fmt.Errorf("repository: get recipe for product %s: %w", productID, err)
	}
	defer rows.Close()

	var results []model.RecipeDetail
	for rows.Next() {
		var row model.RecipeDetail
		if err := rows.Scan(
			&row.IngredientID,
			&row.Name,
			&row.Unit,
			&row.UsageQty,
			&row.TotalDeduct,
		); err != nil {
			return nil, fmt.Errorf("repository: scan recipe row: %w", err)
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// DeductIngredientStock mengurangi stok bahan dengan guard level DB.
// KUNCI: klausa WHERE stock >= $1 mencegah stok negatif secara atomik.
func (r *SaleRepository) DeductIngredientStock(
	ctx context.Context,
	tx pgx.Tx,
	ingredientID uuid.UUID,
	deductQty float64,
) error {
	query := `
		UPDATE ingredients
		SET    stock = stock - $1,
			   updated_at = NOW()
		WHERE  id = $2
		  AND  stock >= $1
		  AND  deleted_at IS NULL
	`
	tag, err := tx.Exec(ctx, query, deductQty, ingredientID)
	if err != nil {
		return fmt.Errorf("repository: deduct stock: %w", err)
	}
	// Jika 0 baris terupdate, berarti stok tidak cukup
	if tag.RowsAffected() == 0 {
		return errors.New("insufficient stock for ingredient: " + ingredientID.String())
	}
	return nil
}

// GetProductPrice mengambil harga produk dari DB di dalam transaksi.
func (r *SaleRepository) GetProductPrice(
	ctx context.Context,
	tx pgx.Tx,
	productID uuid.UUID,
) (float64, string, error) {
	var price float64
	var name string
	query := `SELECT price, name FROM products WHERE id = $1 AND deleted_at IS NULL AND is_available = TRUE`
	err := tx.QueryRow(ctx, query, productID).Scan(&price, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", fmt.Errorf("product not found or unavailable: %s", productID)
		}
		return 0, "", fmt.Errorf("repository: get product price: %w", err)
	}
	return price, name, nil
}

// InsertSale menyimpan header transaksi ke DB.
func (r *SaleRepository) InsertSale(ctx context.Context, tx pgx.Tx, sale *model.Sale) error {
	query := `
		INSERT INTO sales (id, idempotency_key, total_amount, payment_method, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.Exec(ctx, query,
		sale.ID,
		sale.IdempotencyKey,
		sale.TotalAmount,
		sale.PaymentMethod,
		sale.Status,
		sale.Notes,
	)
	if err != nil {
		return fmt.Errorf("repository: insert sale: %w", err)
	}
	return nil
}

// InsertSaleItem menyimpan satu baris detail item ke DB.
func (r *SaleRepository) InsertSaleItem(ctx context.Context, tx pgx.Tx, item *model.SaleItemRecord) error {
	query := `
		INSERT INTO sale_items (id, sale_id, product_id, product_name, quantity, unit_price)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.Exec(ctx, query,
		item.ID,
		item.SaleID,
		item.ProductID,
		item.ProductName,
		item.Quantity,
		item.UnitPrice,
	)
	if err != nil {
		return fmt.Errorf("repository: insert sale item: %w", err)
	}
	return nil
}
