package jwt

import (
	"AuthService/internal/domain/models"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")

func GenerateToken(claims models.TokenClaims, secretKey string, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": claims.Username,
		"userID":   claims.UserID,
		"exp":      time.Now().Add(ttl).Unix(),
		"iat":      time.Now().Unix(),
	})

	return token.SignedString([]byte(secretKey))
}
func ValidateToken(tokenString string, secretKey string) (*models.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["userID"].(float64)

		username := claims["username"].(string)
		return &models.TokenClaims{
			UserID:   int(userID),
			Username: username,
		}, nil
	}
	return nil, ErrInvalidToken
}
