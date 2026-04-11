package handler

// sse_handler.go
// SSE broadcasting ditangani langsung oleh SSEBroker di platform/broker/sse_broker.go.
// SSEBroker mengimplementasikan http.Handler sehingga bisa di-mount langsung ke router:
//
//   mux.Handle("GET /api/events/stock", sseBroker)
//
// File ini disediakan sebagai placeholder sesuai struktur folder.
// Tambahkan auth middleware di sini jika SSE perlu dilindungi dengan JWT.
