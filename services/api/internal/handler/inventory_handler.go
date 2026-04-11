package handler

import (
	"net/http"

	"github.com/ythStevin0/kopi-pos/services/api/internal/model"
	"github.com/ythStevin0/kopi-pos/services/api/internal/service"
)

// InventoryHandler menangani HTTP request terkait inventory bahan baku.
type InventoryHandler struct {
	inventoryService *service.InventoryService
}

// NewInventoryHandler membuat instance InventoryHandler baru.
func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

// ListInventory menangani GET /api/inventory
func (h *InventoryHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	ingredients, err := h.inventoryService.GetAll(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, model.Fail("failed to fetch inventory: "+err.Error()))
		return
	}
	respondJSON(w, http.StatusOK, model.OK("inventory fetched successfully", ingredients))
}

// ListLowStock menangani GET /api/inventory/low-stock
func (h *InventoryHandler) ListLowStock(w http.ResponseWriter, r *http.Request) {
	ingredients, err := h.inventoryService.GetLowStock(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, model.Fail("failed to fetch low stock: "+err.Error()))
		return
	}
	respondJSON(w, http.StatusOK, model.OK("low stock items fetched successfully", ingredients))
}
