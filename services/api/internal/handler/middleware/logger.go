package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter adalah wrapper untuk menangkap status code dari response.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger mencatat setiap request HTTP: method, path, status code, dan durasi.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		log.Printf(
			"[HTTP] %s %s | %d | %s | %s",
			r.Method,
			r.URL.Path,
			rw.statusCode,
			time.Since(start),
			r.RemoteAddr,
		)
	})
}
