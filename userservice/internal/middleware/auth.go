package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/upskill/userservice/internal/utils"
)

type ctxKey string

const userKey ctxKey = "userID"

// AuthMW validates JWT (shared secret)
func AuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		t, err := utils.Parse(token)
		if err != nil || !t.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		claims := t.Claims.(jwt.MapClaims)
		uid := uint(claims["sub"].(float64))
		r = r.WithContext(context.WithValue(r.Context(), userKey, uid))
		next.ServeHTTP(w, r)
	})
}

// UserID extracts id from context
func UserID(r *http.Request) uint { return r.Context().Value(userKey).(uint) }
