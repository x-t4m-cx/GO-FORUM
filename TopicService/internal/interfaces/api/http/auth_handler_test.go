package http

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		requestBody   string
		mockSetup     func(*MockAuthClient)
		expectedCode  int
		expectedError string
	}{
		{
			name:        "Success",
			requestBody: `{"username":"test","password":"pass"}`,
			mockSetup: func(m *MockAuthClient) {
				m.On("Login", mock.Anything, "test", "pass").
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid request",
			requestBody:  `invalid`,
			mockSetup:    func(m *MockAuthClient) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "Service error",
			requestBody: `{"username":"test","password":"pass"}`,
			mockSetup: func(m *MockAuthClient) {
				m.On("Login", mock.Anything, "test", "pass").
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
					}, errors.New("service error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockAuthClient)
			tt.mockSetup(mockClient)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewAuthHandler(mockClient, *logger)

			router := gin.New()
			router.POST("/login", handler.Login)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		setCookie     bool
		mockSetup     func(*MockAuthClient)
		expectedCode  int
		expectedError string
	}{
		{
			name:      "Success with cookie",
			setCookie: true,
			mockSetup: func(m *MockAuthClient) {
				m.On("Logout", mock.Anything).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:      "Success without cookie",
			setCookie: false,
			mockSetup: func(m *MockAuthClient) {
				m.On("Logout", mock.Anything).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:      "Service error",
			setCookie: true,
			mockSetup: func(m *MockAuthClient) {
				m.On("Logout", mock.Anything).
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
					}, errors.New("service error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockAuthClient)
			tt.mockSetup(mockClient)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewAuthHandler(mockClient, *logger)

			router := gin.New()
			router.POST("/logout", handler.Logout)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/logout", nil)
			if tt.setCookie {
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "test_token"})
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		requestBody   string
		mockSetup     func(*MockAuthClient)
		expectedCode  int
		expectedError string
	}{
		{
			name:        "Success",
			requestBody: `{"username":"test","password":"pass"}`,
			mockSetup: func(m *MockAuthClient) {
				m.On("Register", mock.Anything, "test", "pass").
					Return(&http.Response{
						StatusCode: http.StatusCreated,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
					}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Invalid request",
			requestBody:  `invalid`,
			mockSetup:    func(m *MockAuthClient) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "Service error",
			requestBody: `{"username":"test","password":"pass"}`,
			mockSetup: func(m *MockAuthClient) {
				m.On("Register", mock.Anything, "test", "pass").
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
					}, errors.New("service error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockAuthClient)
			tt.mockSetup(mockClient)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewAuthHandler(mockClient, *logger)

			router := gin.New()
			router.POST("/register", handler.Register)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		setCookie     bool
		cookieValue   string
		mockSetup     func(*MockAuthClient)
		expectedCode  int
		expectedError string
	}{
		{
			name:        "Success",
			setCookie:   true,
			cookieValue: "refresh_token_value",
			mockSetup: func(m *MockAuthClient) {
				m.On("Refresh", mock.Anything, "refresh_token_value").
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "No cookie",
			setCookie:     false,
			mockSetup:     func(m *MockAuthClient) {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "http: named cookie not present",
		},
		{
			name:        "Service error",
			setCookie:   true,
			cookieValue: "refresh_token_value",
			mockSetup: func(m *MockAuthClient) {
				m.On("Refresh", mock.Anything, "refresh_token_value").
					Return(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
					}, errors.New("service error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockAuthClient)
			tt.mockSetup(mockClient)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewAuthHandler(mockClient, *logger)

			router := gin.New()
			router.POST("/refresh", handler.Refresh)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/refresh", nil)
			if tt.setCookie {
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tt.cookieValue})
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Verify(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		setHeader     bool
		headerValue   string
		mockSetup     func(*MockAuthClient)
		expectedCode  int
		expectedError string
		checkValid    bool // Новое поле для указания необходимости проверки valid
	}{
		{
			name:        "Success",
			setHeader:   true,
			headerValue: "Bearer valid_token",
			mockSetup: func(m *MockAuthClient) {
				m.On("VerifyToken", mock.Anything, "Bearer valid_token").
					Return("testuser", nil)
			},
			expectedCode: http.StatusOK,
			checkValid:   true,
		},
		{
			name:          "No header",
			setHeader:     false,
			mockSetup:     func(m *MockAuthClient) {},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Authorization header is empty",
			checkValid:    false,
		},
		{
			name:        "Invalid token",
			setHeader:   true,
			headerValue: "Bearer invalid_token",
			mockSetup: func(m *MockAuthClient) {
				m.On("VerifyToken", mock.Anything, "Bearer invalid_token").
					Return("", errors.New("invalid token"))
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "invalid token",
			checkValid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockAuthClient)
			tt.mockSetup(mockClient)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewAuthHandler(mockClient, *logger)

			router := gin.New()
			router.GET("/verify", handler.Verify)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/verify", nil)
			if tt.setHeader {
				req.Header.Set("Authorization", tt.headerValue)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			body := w.Body.String()
			if tt.expectedError != "" {
				assert.Contains(t, body, tt.expectedError)
			}

			if tt.checkValid {
				if tt.expectedCode == http.StatusOK {
					assert.Contains(t, body, `"valid":true`)
				} else {
					assert.Contains(t, body, `"valid":false`)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}
