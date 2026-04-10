package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ythStevin0/kopi-pos/services/api/internal/model"
	"github.com/ythStevin0/kopi-pos/services/api/internal/repository"
)

// ProductService mengandung business logic terkait produk.
type ProductService struct {
	productRepo *repository.ProductRepository
}

// NewProductService membuat instance ProductService baru.
func NewProductService(productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

// GetAll mengembalikan semua produk aktif.
func (s *ProductService) GetAll(ctx context.Context) ([]model.Product, error) {
	products, err := s.productRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: get all products: %w", err)
	}
	return products, nil
}

// GetByID mengembalikan satu produk berdasarkan UUID.
func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: get product by id: %w", err)
	}
	return product, nil
}

// Create membuat produk baru.
func (s *ProductService) Create(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("service: product name is required")
	}
	if req.Price < 0 {
		return nil, fmt.Errorf("service: product price cannot be negative")
	}

	product, err := s.productRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("service: create product: %w", err)
	}
	return product, nil
}

// Delete melakukan soft delete pada produk.
func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.productRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("service: delete product: %w", err)
	}
	return nil
}
