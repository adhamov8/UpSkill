package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/* ---------- секрет ---------- */

var secret = func() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret"
	}
	return []byte(s)
}()

/* ---------- генерация ---------- */

func token(uid uint, role string, ttl time.Duration, typ string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  uid,
		"role": role,
		"type": typ,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func NewAccess(id uint, role string) (string, error) { return token(id, role, time.Hour, "access") }
func NewRefresh(id uint) (string, error)             { return token(id, "", 24*time.Hour, "refresh") }

/* ---------- разбор ---------- */

func Parse(tok string) (*jwt.Token, error) {
	return jwt.Parse(tok, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
}

func ValidateRefresh(tok string) (uint, error) {
	t, err := Parse(tok)
	if err != nil || !t.Valid {
		return 0, errors.New("bad token")
	}
	cl := t.Claims.(jwt.MapClaims)
	if cl["type"] != "refresh" {
		return 0, errors.New("not refresh")
	}
	return uint(cl["sub"].(float64)), nil
}
