package utils

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte(os.Getenv("JWT_SECRET"))

// Parse validates token (sign method + exp)
func Parse(tok string) (*jwt.Token, error) {
	return jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
}
