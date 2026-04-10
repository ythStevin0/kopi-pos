package model

import (
	"time"

	"github.com/google/uuid"
)

// Product adalah domain struct untuk tabel products.
type Product struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Price       float64    `json:"price" db:"price"`
	Category    string     `json:"category" db:"category"`
	ImageURL    string     `json:"image_url,omitempty" db:"image_url"`
	IsAvailable bool       `json:"is_available" db:"is_available"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateProductRequest adalah DTO untuk membuat produk baru.
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	ImageURL    string  `json:"image_url,omitempty"`
}
