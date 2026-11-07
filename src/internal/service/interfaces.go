package service

import (
	"github.com/golang-jwt/jwt/v5"
)

type TokenProvider interface {
	GenerateAccessToken(user User) (string, error)
	GenerateRefreshToken(user User) (string, error)

	ValidateAccessToken(token string) error
	ValidateRefreshToken(token string) error

	GetUserID(token string) (string, error)
	GetUserIDFromRefreshToken(tokenString string) (string, error)
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
