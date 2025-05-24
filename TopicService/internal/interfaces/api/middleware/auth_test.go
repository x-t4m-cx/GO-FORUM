package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Login(ctx context.Context, username, password string) (*http.Response, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockAuthClient) Logout(ctx context.Context) (*http.Response, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockAuthClient) Register(ctx context.Context, username, password string) (*http.Response, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockAuthClient) Refresh(ctx context.Context, refreshToken string) (*http.Response, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockAuthClient) VerifyToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockAuthClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestAuthMiddleware_Auth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful authentication with valid token", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		middleware := NewAuthMiddleware(mockClient, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))

		mockClient.On("VerifyToken", mock.Anything, "valid-token").Return("testuser", nil)

		router := gin.New()
		router.Use(middleware.Auth())
		router.GET("/test", func(c *gin.Context) {
			username, _ := c.Get("username")
			c.JSON(http.StatusOK, gin.H{"username": username})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "testuser")
		mockClient.AssertExpectations(t)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		middleware := NewAuthMiddleware(mockClient, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))

		router := gin.New()
		router.Use(middleware.Auth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Authorization header is required")
		mockClient.AssertNotCalled(t, "VerifyToken")
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		middleware := NewAuthMiddleware(mockClient, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))

		router := gin.New()
		router.Use(middleware.Auth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Invalid authorization header format")
		mockClient.AssertNotCalled(t, "VerifyToken")
	})

	t.Run("empty token", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		middleware := NewAuthMiddleware(mockClient, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))

		router := gin.New()
		router.Use(middleware.Auth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer ")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Empty token provided")
		mockClient.AssertNotCalled(t, "VerifyToken")
	})

	t.Run("invalid token but successful refresh", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		middleware := NewAuthMiddleware(mockClient, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))

		// First VerifyToken call fails
		mockClient.On("VerifyToken", mock.Anything, "expired-token").Return("", errors.New("token expired"))

		// Mock refresh response
		refreshResponse := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
		}
		refreshResponse.Header.Set("Authorization", "Bearer new-token")
		mockClient.On("Refresh", mock.Anything, "refresh-token").Return(refreshResponse, nil)

		// Second VerifyToken call succeeds with new token
		mockClient.On("VerifyToken", mock.Anything, "new-token").Return("testuser", nil)

		router := gin.New()
		router.Use(middleware.Auth())
		router.GET("/test", func(c *gin.Context) {
			username, _ := c.Get("username")
			c.JSON(http.StatusOK, gin.H{"username": username})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer expired-token")
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token"})
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "testuser")
		mockClient.AssertExpectations(t)
	})

	t.Run("missing refresh token when access token is invalid", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		middleware := NewAuthMiddleware(mockClient, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))

		mockClient.On("VerifyToken", mock.Anything, "expired-token").Return("", errors.New("token expired"))

		router := gin.New()
		router.Use(middleware.Auth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer expired-token")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Refresh token required")
		mockClient.AssertExpectations(t)
	})

	t.Run("refresh token fails", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		middleware := NewAuthMiddleware(mockClient, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))

		mockClient.On("VerifyToken", mock.Anything, "expired-token").Return("", errors.New("token expired"))
		mockClient.On("Refresh", mock.Anything, "invalid-refresh-token").Return(
			&http.Response{StatusCode: http.StatusUnauthorized},
			errors.New("refresh failed"),
		)

		router := gin.New()
		router.Use(middleware.Auth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer expired-token")
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-refresh-token"})
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Failed to refresh token")
		mockClient.AssertExpectations(t)
	})

	t.Run("refreshed token is invalid", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		middleware := NewAuthMiddleware(mockClient, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))

		// First VerifyToken call fails
		mockClient.On("VerifyToken", mock.Anything, "expired-token").Return("", errors.New("token expired"))

		// Mock refresh response
		refreshResponse := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
		}
		refreshResponse.Header.Set("Authorization", "Bearer invalid-new-token")
		mockClient.On("Refresh", mock.Anything, "refresh-token").Return(refreshResponse, nil)

		// Second VerifyToken call fails with new token
		mockClient.On("VerifyToken", mock.Anything, "invalid-new-token").Return("", errors.New("invalid token"))

		router := gin.New()
		router.Use(middleware.Auth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer expired-token")
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token"})
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Contains(t, resp.Body.String(), "Refreshed token is invalid")
		mockClient.AssertExpectations(t)
	})
}
