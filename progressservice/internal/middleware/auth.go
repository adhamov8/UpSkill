package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/upskill/progressservice/internal/utils"
)

type ctxKey string

const userKey ctxKey = "uid"

func AuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		t, err := utils.Parse(tok)
		if err != nil || !t.Valid {
			http.Error(w, "unauthorized", 401)
			return
		}
		uid := uint(t.Claims.(jwt.MapClaims)["sub"].(float64))
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, uid)))
	})
}

func UserID(r *http.Request) uint { return r.Context().Value(userKey).(uint) }
