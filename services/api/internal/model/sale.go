package model

import (
	"time"

	"github.com/google/uuid"
)

// SaleItem adalah satu item dalam request transaksi.
type SaleItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// ProcessSaleRequest adalah DTO dari client.
type ProcessSaleRequest struct {
	IdempotencyKey uuid.UUID  `json:"idempotency_key"`
	Items          []SaleItem `json:"items"`
	PaymentMethod  string     `json:"payment_method"`
	Notes          string     `json:"notes,omitempty"`
}

// Sale adalah domain struct untuk tabel sales.
type Sale struct {
	ID             uuid.UUID `json:"id"`
	IdempotencyKey uuid.UUID `json:"idempotency_key"`
	TotalAmount    float64   `json:"total_amount"`
	PaymentMethod  string    `json:"payment_method"`
	Status         string    `json:"status"`
	Notes          string    `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// SaleItemRecord adalah domain struct untuk tabel sale_items.
type SaleItemRecord struct {
	ID          uuid.UUID `json:"id"`
	SaleID      uuid.UUID `json:"sale_id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Subtotal    float64   `json:"subtotal"`
}
