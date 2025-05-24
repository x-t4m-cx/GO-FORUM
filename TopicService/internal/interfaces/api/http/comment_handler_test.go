package http

import (
	"TopicService/internal/domain/models"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCommentUseCase struct {
	mock.Mock
}

func (m *MockCommentUseCase) CreateComment(ctx context.Context, comment *models.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentUseCase) GetAllComments(ctx context.Context, topicID int) ([]*models.Comment, error) {
	args := m.Called(ctx, topicID)
	return args.Get(0).([]*models.Comment), args.Error(1)
}

func (m *MockCommentUseCase) GetComment(ctx context.Context, id int) (*models.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Comment), args.Error(1)
}

func (m *MockCommentUseCase) UpdateComment(ctx context.Context, comment *models.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentUseCase) DeleteComment(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCommentHandler_CreateComment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockCommentUseCase)
		setupAuth      func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			requestBody: `{
				"topic_id": "1",
				"content": "Test comment content"
			}`,
			mockSetup: func(m *MockCommentUseCase) {
				m.On("CreateComment", mock.Anything, mock.MatchedBy(func(comment *models.Comment) bool {
					return comment.TopicID == 1 && comment.Content == "Test comment content"
				})).Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"topic_id":1,"username":"testuser"`,
		},
		{
			name:           "Unauthorized",
			requestBody:    `{}`,
			mockSetup:      func(m *MockCommentUseCase) {},
			setupAuth:      func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `"error":"unauthorized"`,
		},
		{
			name: "Invalid Request Body",
			requestBody: `{
				"topic_id": "1",
				"content": "Test comment content"
			`,
			mockSetup: func(m *MockCommentUseCase) {},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"unexpected EOF"`,
		},
		{
			name: "Invalid Topic ID",
			requestBody: `{
				"topic_id": "abc",
				"content": "Test comment content"
			}`,
			mockSetup: func(m *MockCommentUseCase) {},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"strconv.Atoi: parsing \"abc\": invalid syntax"`,
		},
		{
			name: "Service Error",
			requestBody: `{
				"topic_id": "1",
				"content": "Test comment content"
			}`,
			mockSetup: func(m *MockCommentUseCase) {
				m.On("CreateComment", mock.Anything, mock.AnythingOfType("*models.Comment")).
					Return(errors.New("service error"))
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"service error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			mockUseCase := new(MockCommentUseCase)
			tt.mockSetup(mockUseCase)

			handler := NewCommentHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.Use(tt.setupAuth)
			router.POST("/comments/", handler.CreateComment)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/comments/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestCommentHandler_GetAll(t *testing.T) {
	now := time.Now()
	testComments := []*models.Comment{
		{Id: 1, TopicID: 1, Username: "user1", Content: "Content 1", CreatedAt: now},
		{Id: 2, TopicID: 1, Username: "user2", Content: "Content 2", CreatedAt: now},
	}

	tests := []struct {
		name           string
		topicID        string
		mockSetup      func(*MockCommentUseCase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			topicID: "1",
			mockSetup: func(m *MockCommentUseCase) {
				m.On("GetAllComments", mock.Anything, 1).
					Return(testComments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"topic_id":1`,
		},
		{
			name:           "Invalid Topic ID",
			topicID:        "abc",
			mockSetup:      func(m *MockCommentUseCase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"strconv.Atoi: parsing \"abc\": invalid syntax"`,
		},
		{
			name:    "Service Error",
			topicID: "1",
			mockSetup: func(m *MockCommentUseCase) {
				m.On("GetAllComments", mock.Anything, 1).
					Return([]*models.Comment{}, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"service error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockCommentUseCase)
			tt.mockSetup(mockUseCase)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewCommentHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.GET("/topics/comments/:topic_id", handler.GetAll)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/topics/comments/"+tt.topicID, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestCommentHandler_GetComment(t *testing.T) {
	now := time.Now()
	testComment := &models.Comment{
		Id:        1,
		TopicID:   1,
		Username:  "testuser",
		Content:   "Test content",
		CreatedAt: now,
	}

	tests := []struct {
		name           string
		commentID      string
		mockSetup      func(*MockCommentUseCase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "Success",
			commentID: "1",
			mockSetup: func(m *MockCommentUseCase) {
				m.On("GetComment", mock.Anything, 1).
					Return(testComment, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"id":1,"topic_id":1`,
		},
		{
			name:           "Invalid ID",
			commentID:      "abc",
			mockSetup:      func(m *MockCommentUseCase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"strconv.Atoi: parsing \"abc\": invalid syntax"`,
		},
		{
			name:      "Service Error",
			commentID: "1",
			mockSetup: func(m *MockCommentUseCase) {
				m.On("GetComment", mock.Anything, 1).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"service error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockCommentUseCase)
			tt.mockSetup(mockUseCase)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewCommentHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.GET("/comments/:id", handler.GetComment)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/comments/"+tt.commentID, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestCommentHandler_UpdateComment(t *testing.T) {
	tests := []struct {
		name           string
		commentID      string
		requestBody    string
		mockSetup      func(*MockCommentUseCase)
		setupAuth      func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "Success - Author",
			commentID: "1",
			requestBody: `{
				"content": "Updated content"
			}`,
			mockSetup: func(m *MockCommentUseCase) {
				m.On("GetComment", mock.Anything, 1).
					Return(&models.Comment{Id: 1, Username: "testuser"}, nil)
				m.On("UpdateComment", mock.Anything, mock.MatchedBy(func(comment *models.Comment) bool {
					return comment.Id == 1 && comment.Content == "Updated content"
				})).Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"Comment updated successfully"`,
		},
		{
			name:      "Success - Admin",
			commentID: "1",
			requestBody: `{
				"content": "Updated content"
			}`,
			mockSetup: func(m *MockCommentUseCase) {
				m.On("GetComment", mock.Anything, 1).
					Return(&models.Comment{Id: 1, Username: "otheruser"}, nil)
				m.On("UpdateComment", mock.Anything, mock.MatchedBy(func(comment *models.Comment) bool {
					return comment.Id == 1 && comment.Content == "Updated content"
				})).Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "admin")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"Comment updated successfully"`,
		},
		{
			name:        "Invalid ID",
			commentID:   "abc",
			requestBody: `{}`,
			mockSetup:   func(m *MockCommentUseCase) {},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"Invalid comment ID"`,
		},
		{
			name:      "Invalid Request Body",
			commentID: "1",
			requestBody: `{
				"content": "Updated content"
			`,
			mockSetup: func(m *MockCommentUseCase) {},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"unexpected EOF"`,
		},
		{
			name:      "Unauthorized - Not Author",
			commentID: "1",
			requestBody: `{
				"content": "Updated content"
			}`,
			mockSetup: func(m *MockCommentUseCase) {
				m.On("GetComment", mock.Anything, 1).
					Return(&models.Comment{Id: 1, Username: "otheruser"}, nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `"error":"forbidden"`,
		},
		{
			name:      "Service Error - GetComment",
			commentID: "1",
			requestBody: `{
				"content": "Updated content"
			}`,
			mockSetup: func(m *MockCommentUseCase) {
				m.On("GetComment", mock.Anything, 1).
					Return(nil, errors.New("service error"))
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"service error"`,
		},
		{
			name:      "Service Error - UpdateComment",
			commentID: "1",
			requestBody: `{
				"content": "Updated content"
			}`,
			mockSetup: func(m *MockCommentUseCase) {
				m.On("GetComment", mock.Anything, 1).
					Return(&models.Comment{Id: 1, Username: "testuser"}, nil)
				m.On("UpdateComment", mock.Anything, mock.AnythingOfType("*models.Comment")).
					Return(errors.New("service error"))
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"service error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockCommentUseCase)
			tt.mockSetup(mockUseCase)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewCommentHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.Use(tt.setupAuth)
			router.PUT("/comments/:id", handler.UpdateComment)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/comments/"+tt.commentID, strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	tests := []struct {
		name           string
		commentID      string
		mockSetup      func(*MockCommentUseCase)
		setupAuth      func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "Success",
			commentID: "1",
			mockSetup: func(m *MockCommentUseCase) {
				m.On("DeleteComment", mock.Anything, 1).
					Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"Comment deleted successfully!"`,
		},
		{
			name:      "Invalid ID",
			commentID: "abc",
			mockSetup: func(m *MockCommentUseCase) {},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"strconv.Atoi: parsing \"abc\": invalid syntax"`,
		},
		{
			name:      "Service Error - DeleteComment",
			commentID: "1",
			mockSetup: func(m *MockCommentUseCase) {
				m.On("DeleteComment", mock.Anything, 1).
					Return(errors.New("service error"))
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"service error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockCommentUseCase)
			tt.mockSetup(mockUseCase)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewCommentHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.Use(tt.setupAuth)
			router.DELETE("/comments/:id", handler.DeleteComment)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/comments/"+tt.commentID, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			mockUseCase.AssertExpectations(t)
		})
	}
}
