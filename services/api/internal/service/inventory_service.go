package service

import (
	"context"
	"fmt"

	"github.com/ythStevin0/kopi-pos/services/api/internal/model"
	"github.com/ythStevin0/kopi-pos/services/api/internal/repository"
)

// InventoryService mengandung business logic terkait inventory bahan baku.
type InventoryService struct {
	ingredientRepo *repository.IngredientRepository
}

// NewInventoryService membuat instance InventoryService baru.
func NewInventoryService(ingredientRepo *repository.IngredientRepository) *InventoryService {
	return &InventoryService{ingredientRepo: ingredientRepo}
}

// GetAll mengembalikan semua ingredient aktif.
func (s *InventoryService) GetAll(ctx context.Context) ([]model.Ingredient, error) {
	ingredients, err := s.ingredientRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: get all ingredients: %w", err)
	}
	return ingredients, nil
}

// GetLowStock mengembalikan ingredient yang stoknya di bawah min_stock.
func (s *InventoryService) GetLowStock(ctx context.Context) ([]model.Ingredient, error) {
	all, err := s.ingredientRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: get low stock: %w", err)
	}

	var lowStock []model.Ingredient
	for _, ing := range all {
		if ing.Stock <= ing.MinStock {
			lowStock = append(lowStock, ing)
		}
	}
	return lowStock, nil
}
