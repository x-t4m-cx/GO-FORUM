package http

import (
	"TopicService/internal/domain/models"
	"TopicService/internal/usecases"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http"
)

type AuthHandler struct {
	client usecases.GRPCClientInterface
	logger slog.Logger
}

func NewAuthHandler(client usecases.GRPCClientInterface, logger slog.Logger) *AuthHandler {
	return &AuthHandler{client: client, logger: logger}
}

// Login handles user login
// @Summary User login
// @Description Authenticates user and returns JWT tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Returns access and refresh tokens"
// @Failure 400 {object} models.ErrorResponse "Invalid request format"
// @Failure 401 {object} models.ErrorResponse "Invalid credentials"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var request models.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.Login(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		c.JSON(httpStatusCodeFromError(err), models.ErrorResponse{Error: err.Error()})
		return
	}

	copyHeadersAndCookies(c, resp)
	copyResponseBody(c, resp)
}

// Logout handles user logout
// @Summary User logout
// @Description Invalidates refresh token and clears cookies
// @Tags Authentication
// @Produce json
// @Success 200 "Logout successful"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	if token, err := c.Cookie("refresh_token"); err == nil {
		ctx := context.WithValue(c.Request.Context(), "refresh_token", token)
		c.Request = c.Request.WithContext(ctx)
	}

	resp, err := h.client.Logout(c.Request.Context())
	if err != nil {
		c.JSON(httpStatusCodeFromError(err), models.ErrorResponse{Error: err.Error()})
		return
	}

	copyHeadersAndCookies(c, resp)
	copyResponseBody(c, resp)
}

// Register handles new user registration
// @Summary Register new user
// @Description Creates new user account and returns JWT tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration data"
// @Success 201 {object} map[string]interface{} "Returns access and refresh tokens"
// @Failure 400 {object} models.ErrorResponse "Invalid request format"
// @Failure 409 {object} models.ErrorResponse "User already exists"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var request models.RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.client.Register(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		c.JSON(httpStatusCodeFromError(err), models.ErrorResponse{Error: err.Error()})
		return
	}
	copyHeadersAndCookies(c, resp)
	copyResponseBody(c, resp)
}

// Refresh generates new access token
// @Summary Refresh access token
// @Description Generates new access token using refresh token
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]interface{} "Returns new access token"
// @Failure 400 {object} models.ErrorResponse "Refresh token missing or invalid"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.client.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(httpStatusCodeFromError(err), models.ErrorResponse{Error: err.Error()})
		return
	}

	copyHeadersAndCookies(c, resp)
	copyResponseBody(c, resp)
}

// Verify checks token validity
// @Summary Verify access token
// @Description Verifies JWT token and returns username if valid
// @Tags Authentication
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Success 200 {object} map[string]interface{} "Returns username and validation status"
// @Failure 401 {object} models.ErrorResponse "Token is invalid or expired"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/verify [get]
func (h *AuthHandler) Verify(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authorization header is empty"})
		return
	}
	username, err := h.client.VerifyToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username": username,
		"valid":    true,
	})
}

func httpStatusCodeFromError(err error) int {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func copyHeadersAndCookies(c *gin.Context, resp *http.Response) {
	for name, values := range resp.Header {
		c.Header(name, values[0])
	}
}

func copyResponseBody(c *gin.Context, resp *http.Response) {
	c.Status(resp.StatusCode)
	if resp.Body != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			return
		}
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}
