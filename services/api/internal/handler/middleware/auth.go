package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ythStevin0/kopi-pos/services/api/internal/model"
)

// contextKey adalah type privat untuk menghindari konflik key di context.
type contextKey string

// UserIDKey adalah key untuk menyimpan user ID di request context.
const UserIDKey contextKey = "userID"

// JWTAuth memvalidasi JWT Bearer token dari header Authorization.
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(model.Fail("missing or invalid authorization header"))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(model.Fail("invalid or expired token"))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(model.Fail("invalid token claims"))
			return
		}

		// Inject userID ke context agar handler bisa menggunakannya
		ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
