package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUser struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
}

func BuildJWTClaims(user JWTUser, expireDuration time.Duration) jwt.MapClaims {
	return jwt.MapClaims{
		"userId": user.UserID,
		"name":   user.Name,
		"email":  user.Email,
		"phone":  user.Phone,
		"exp":    jwt.NewNumericDate(time.Now().Add(expireDuration)),
	}
}
