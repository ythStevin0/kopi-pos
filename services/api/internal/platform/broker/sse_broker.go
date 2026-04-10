package broker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// StockUpdateEvent adalah payload yang di-broadcast ke semua client SSE.
type StockUpdateEvent struct {
	IngredientID string  `json:"ingredient_id"`
	Name         string  `json:"name"`
	NewStock     float64 `json:"new_stock"`
}

// SSEBroker mengelola koneksi SSE dari semua client frontend.
type SSEBroker struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

// NewSSEBroker membuat instance SSEBroker baru.
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		clients: make(map[chan []byte]struct{}),
	}
}

// Subscribe mendaftarkan channel baru untuk seorang client.
func (b *SSEBroker) Subscribe() chan []byte {
	ch := make(chan []byte, 8)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	log.Printf("[SSE] Client connected. Total: %d", len(b.clients))
	return ch
}

// Unsubscribe membersihkan channel saat client disconnect.
func (b *SSEBroker) Unsubscribe(ch chan []byte) {
	b.mu.Lock()
	delete(b.clients, ch)
	close(ch)
	b.mu.Unlock()
	log.Printf("[SSE] Client disconnected. Total: %d", len(b.clients))
}

// Broadcast mengirim event ke SEMUA client yang terhubung.
func (b *SSEBroker) Broadcast(event StockUpdateEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("[SSE] Failed to marshal event: %v", err)
		return
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- payload:
		default:
			// Jika channel penuh, skip (non-blocking) agar tidak memblokir broadcast
		}
	}
}

// ServeHTTP adalah HTTP handler untuk endpoint SSE (/api/events/stock).
func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // penting untuk Nginx/Railway proxy

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	for {
		select {
		case payload, open := <-ch:
			if !open {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
