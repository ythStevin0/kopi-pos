package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORS hanya mengizinkan request dari domain Vercel yang dikonfigurasi via ENV.
// Set ALLOWED_ORIGINS="https://kopipos.vercel.app,https://kopipos-preview.vercel.app"
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
		origin := r.Header.Get("Origin")

		allowed := false
		for _, o := range allowedOrigins {
			if strings.TrimSpace(o) == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin") // penting untuk caching yang benar
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Idempotency-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
