package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := tokenFromRequest(r)
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "spaceedu-super-secret-jwt-key-change-in-production"
		}

		t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !t.Valid {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"invalid claims"}`, http.StatusUnauthorized)
			return
		}

		userID, _ := claims["sub"].(string)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func tokenFromRequest(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if c, err := r.Cookie("token"); err == nil {
		return c.Value
	}
	return ""
}

func GetUserID(r *http.Request) string {
	v, _ := r.Context().Value(UserIDKey).(string)
	return v
}
