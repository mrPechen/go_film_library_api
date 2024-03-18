package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"src/src/database"
	"time"
)

// Функция для создания access token
func CreateAccessToken(user *database.User, signingKey []byte, tokenExpires time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(tokenExpires).Unix(),
	})
	return token.SignedString(signingKey)
}

// Функция для создания refresh token
func CreateRefreshToken(user *database.User, refreshKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
	})
	return token.SignedString(refreshKey)
}
