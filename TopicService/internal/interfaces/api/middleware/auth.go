package middleware

import (
	"TopicService/internal/usecases"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	authClient usecases.GRPCClientInterface
}

func NewAuthMiddleware(authClient usecases.GRPCClientInterface, logger slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient}
}

func (m *AuthMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			return
		}
		token := headerParts[1]
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Empty token provided",
			})
			return
		}

		username, err := m.authClient.VerifyToken(c.Request.Context(), token)
		if err == nil {
			c.Set("username", username)
			c.Next()
			return
		}

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Refresh token required",
			})
			return
		}

		resp, err := m.authClient.Refresh(c.Request.Context(), refreshToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Failed to refresh token: " + err.Error(),
			})
			return
		}

		if resp.StatusCode != http.StatusOK {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token refresh failed",
			})
			return
		}

		newAccessToken := resp.Header.Get("Authorization")
		if newAccessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "New access token not provided",
			})
			return
		}

		c.Request.Header.Set("Authorization", newAccessToken)

		username, err = m.authClient.VerifyToken(
			c.Request.Context(),
			strings.TrimPrefix(newAccessToken, "Bearer "),
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Refreshed token is invalid: " + err.Error(),
			})
			return
		}

		c.Set("username", username)
		c.Next()
	}
}
