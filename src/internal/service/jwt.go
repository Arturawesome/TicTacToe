package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtProvider struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewJwtProvider(accessSecret string, refreshSecret string) *JwtProvider {
	return &JwtProvider{accessSecret: []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
	}
}

func (provider *JwtProvider) GenerateAccessToken(user User) (string, error) {

	claims := UserClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(provider.accessSecret)
}
func (provider *JwtProvider) GenerateRefreshToken(user User) (string, error) {
	claims := UserClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(provider.refreshSecret)
}

func (provider *JwtProvider) ValidateAccessToken(tokenString string) error {
	// func to method validate
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return provider.accessSecret, nil
	}

	claims := &UserClaims{}
	jwtToken, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)

	if err != nil {
		return err
	}
	if !jwtToken.Valid {
		return jwt.ErrTokenInvalidClaims
	}
	return nil
}
func (provider *JwtProvider) ValidateRefreshToken(tokenString string) error {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return provider.refreshSecret, nil
	}
	claims := &UserClaims{}
	jwtToken, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return err
	}
	if !jwtToken.Valid {
		return jwt.ErrTokenInvalidClaims
	}
	return nil

}

func (provider *JwtProvider) GetUserID(tokenString string) (string, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return provider.accessSecret, nil
	}
	claims := &UserClaims{}
	jwtToken, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return "", err
	}
	if !jwtToken.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}
	return claims.UserID, nil
}
func (provider *JwtProvider) GetUserIDFromRefreshToken(tokenString string) (string, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return provider.refreshSecret, nil
	}
	claims := &UserClaims{}
	jwtToken, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return "", err
	}
	if !jwtToken.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}
	return claims.UserID, nil
}
