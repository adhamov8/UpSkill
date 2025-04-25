package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/upskill/authservice/internal/models"
	"github.com/upskill/authservice/internal/utils"
)

type ctxKey string

const userCtx ctxKey = "user"

func AuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		t, _ := utils.Parse(tok)
		if t == nil || !t.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userCtx, t.Claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoleMW(allowed ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(userCtx).(jwt.MapClaims)["role"].(string)
			for _, a := range allowed {
				if role == string(a) {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}
