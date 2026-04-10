package model

import (
	"time"

	"github.com/google/uuid"
)

// Ingredient adalah domain struct untuk tabel ingredients (bahan baku).
type Ingredient struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Unit      string     `json:"unit" db:"unit"`
	Stock     float64    `json:"stock" db:"stock"`
	MinStock  float64    `json:"min_stock" db:"min_stock"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Recipe adalah junction antara product dan ingredients.
type Recipe struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ProductID     uuid.UUID `json:"product_id" db:"product_id"`
	IngredientID  uuid.UUID `json:"ingredient_id" db:"ingredient_id"`
	UsageQuantity float64   `json:"usage_quantity" db:"usage_quantity"`
}

// RecipeDetail adalah recipe dengan info ingredient lengkap.
type RecipeDetail struct {
	IngredientID uuid.UUID `json:"ingredient_id"`
	Name         string    `json:"name"`
	Unit         string    `json:"unit"`
	UsageQty     float64   `json:"usage_qty"`
	TotalDeduct  float64   `json:"total_deduct"`
}
