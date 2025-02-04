package auth

import (
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func ValidateName(name string) bool {
	if len(name) < 2 || len(name) > 30 {
		return false
	}
	matched, _ := regexp.MatchString(`^[A-Za-zА-Яа-я]+$`, name)
	return matched
}

func ValidateEmail(email string) bool {
	matched, _ := regexp.MatchString(`^[^@\s]+@[^@\s]+\.[^@\s]+$`, email)
	return matched
}

func ValidatePassword(pass string) bool {
	if len(pass) < 8 {
		return false
	}
	digitMatch, _ := regexp.MatchString(`[0-9]`, pass)
	if !digitMatch {
		return false
	}
	upperMatch, _ := regexp.MatchString(`[A-Z]`, pass)
	if !upperMatch {
		return false
	}
	return true
}

func GenerateJWT(userID int, secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
