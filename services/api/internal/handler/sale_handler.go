package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ythStevin0/kopi-pos/services/api/internal/model"
	"github.com/ythStevin0/kopi-pos/services/api/internal/service"
)

// SaleHandler menangani HTTP request terkait transaksi penjualan.
type SaleHandler struct {
	saleService *service.SaleService
}

// NewSaleHandler membuat instance SaleHandler baru.
func NewSaleHandler(saleService *service.SaleService) *SaleHandler {
	return &SaleHandler{saleService: saleService}
}

// ProcessSale menangani POST /api/sales
func (h *SaleHandler) ProcessSale(w http.ResponseWriter, r *http.Request) {
	var req model.ProcessSaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, model.Fail("invalid request body: "+err.Error()))
		return
	}

	// Validasi field wajib
	if req.IdempotencyKey == uuid.Nil {
		respondJSON(w, http.StatusBadRequest, model.Fail("idempotency_key is required"))
		return
	}
	if len(req.Items) == 0 {
		respondJSON(w, http.StatusBadRequest, model.Fail("items cannot be empty"))
		return
	}
	if req.PaymentMethod == "" {
		respondJSON(w, http.StatusBadRequest, model.Fail("payment_method is required"))
		return
	}

	sale, err := h.saleService.ProcessSale(r.Context(), req)
	if err != nil {
		// Duplikat transaksi bukan server error — kembalikan 200 agar client tidak retry
		if strings.HasPrefix(err.Error(), "duplicate_transaction") {
			respondJSON(w, http.StatusOK, model.Fail("transaction already processed (idempotent)"))
			return
		}
		// Stok tidak cukup = konflik bisnis
		if strings.Contains(err.Error(), "insufficient stock") {
			respondJSON(w, http.StatusConflict, model.Fail(err.Error()))
			return
		}
		respondJSON(w, http.StatusInternalServerError, model.Fail("failed to process sale: "+err.Error()))
		return
	}

	respondJSON(w, http.StatusCreated, model.OK("sale processed successfully", sale))
}

// respondJSON adalah helper untuk menulis JSON response secara konsisten.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
