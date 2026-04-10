package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ythStevin0/kopi-pos/services/api/internal/model"
	"github.com/ythStevin0/kopi-pos/services/api/internal/platform/broker"
	"github.com/ythStevin0/kopi-pos/services/api/internal/repository"
)

// SaleService mengandung semua business logic terkait transaksi penjualan.
type SaleService struct {
	db        *pgxpool.Pool
	saleRepo  *repository.SaleRepository
	sseBroker *broker.SSEBroker
}

// NewSaleService membuat instance SaleService baru dengan dependency injection.
func NewSaleService(
	db *pgxpool.Pool,
	saleRepo *repository.SaleRepository,
	sseBroker *broker.SSEBroker,
) *SaleService {
	return &SaleService{db: db, saleRepo: saleRepo, sseBroker: sseBroker}
}

// ProcessSale adalah inti bisnis KopiPOS: validasi → transaksi atomik → broadcast.
func (s *SaleService) ProcessSale(ctx context.Context, req model.ProcessSaleRequest) (*model.Sale, error) {
	// ─── STEP 1: IDEMPOTENCY CHECK ────────────────────────────────────────────
	// Cek di luar transaksi agar lebih cepat (read-only query).
	exists, err := s.saleRepo.IdempotencyKeyExists(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("service: idempotency check failed: %w", err)
	}
	if exists {
		// Bukan error fatal — kembalikan konfirmasi tanpa proses ulang.
		return nil, fmt.Errorf("duplicate_transaction: idempotency key already used")
	}

	// ─── STEP 2: MULAI ACID TRANSACTION ──────────────────────────────────────
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("service: begin transaction: %w", err)
	}

	// Pastikan rollback otomatis jika terjadi error
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// ─── STEP 3: PROSES SETIAP ITEM ──────────────────────────────────────────
	saleID := uuid.New()
	var totalAmount float64
	var stockEvents []broker.StockUpdateEvent

	for _, item := range req.Items {
		// Ambil harga dan nama produk (snapshot untuk sale_items)
		price, productName, err := s.saleRepo.GetProductPrice(ctx, tx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("service: get product %s: %w", item.ProductID, err)
		}
		totalAmount += price * float64(item.Quantity)

		// Simpan sale item record
		saleItem := &model.SaleItemRecord{
			ID:          uuid.New(),
			SaleID:      saleID,
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			UnitPrice:   price,
		}
		if err = s.saleRepo.InsertSaleItem(ctx, tx, saleItem); err != nil {
			return nil, fmt.Errorf("service: insert sale item: %w", err)
		}

		// Ambil resep produk (bahan yang diperlukan)
		recipes, err := s.saleRepo.GetRecipeByProductID(ctx, tx, item.ProductID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("service: get recipe for product %s: %w", item.ProductID, err)
		}

		// ─── STEP 4: ATOMIC STOCK DEDUCTION ──────────────────────────────────
		for _, recipe := range recipes {
			if err = s.saleRepo.DeductIngredientStock(ctx, tx, recipe.IngredientID, recipe.TotalDeduct); err != nil {
				return nil, fmt.Errorf("service: stock deduction failed: %w", err)
			}
			// Kumpulkan event untuk SSE broadcast setelah commit
			stockEvents = append(stockEvents, broker.StockUpdateEvent{
				IngredientID: recipe.IngredientID.String(),
				Name:         recipe.Name,
			})
		}
	}

	// ─── STEP 5: INSERT SALE RECORD ──────────────────────────────────────────
	sale := &model.Sale{
		ID:             saleID,
		IdempotencyKey: req.IdempotencyKey,
		TotalAmount:    totalAmount,
		PaymentMethod:  req.PaymentMethod,
		Status:         "completed",
		Notes:          req.Notes,
		CreatedAt:      time.Now(),
	}
	if err = s.saleRepo.InsertSale(ctx, tx, sale); err != nil {
		return nil, err
	}

	// ─── STEP 6: COMMIT ───────────────────────────────────────────────────────
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("service: commit transaction: %w", err)
	}

	// ─── STEP 7: BROADCAST SSE (post-commit, non-blocking) ───────────────────
	go func() {
		for _, event := range stockEvents {
			s.sseBroker.Broadcast(event)
		}
	}()

	return sale, nil
}
