package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/ythStevin0/kopi-pos/services/api/internal/model"
	"github.com/ythStevin0/kopi-pos/services/api/internal/service"
)

// ProductHandler menangani HTTP request terkait produk.
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler membuat instance ProductHandler baru.
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// ListProducts menangani GET /api/products
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productService.GetAll(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, model.Fail("failed to fetch products: "+err.Error()))
		return
	}
	respondJSON(w, http.StatusOK, model.OK("products fetched successfully", products))
}

// CreateProduct menangani POST /api/products
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, model.Fail("invalid request body: "+err.Error()))
		return
	}

	product, err := h.productService.Create(r.Context(), &req)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, model.Fail(err.Error()))
		return
	}
	respondJSON(w, http.StatusCreated, model.OK("product created successfully", product))
}

// DeleteProduct menangani DELETE /api/products/{id}
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari path: /api/products/{id}
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, model.Fail("invalid product id"))
		return
	}

	if err := h.productService.Delete(r.Context(), id); err != nil {
		respondJSON(w, http.StatusNotFound, model.Fail(err.Error()))
		return
	}
	respondJSON(w, http.StatusOK, model.OK("product deleted successfully", nil))
}
