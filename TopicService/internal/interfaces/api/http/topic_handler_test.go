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

type MockTopicUseCase struct {
	mock.Mock
}

func (m *MockTopicUseCase) CreateTopic(ctx context.Context, topic *models.Topic) error {
	args := m.Called(ctx, topic)
	return args.Error(0)
}

func (m *MockTopicUseCase) GetAllTopics(ctx context.Context) ([]*models.Topic, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Topic), args.Error(1)
}

func (m *MockTopicUseCase) GetTopic(ctx context.Context, id int) (*models.Topic, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Topic), args.Error(1)
}

func (m *MockTopicUseCase) UpdateTopic(ctx context.Context, topic *models.Topic) error {
	args := m.Called(ctx, topic)
	return args.Error(0)
}

func (m *MockTopicUseCase) DeleteTopic(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTopicHandler_CreateTopic(t *testing.T) {

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockTopicUseCase)
		setupAuth      func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			requestBody: `{
				"title": "Test Topic",
				"content": "Test Content"
			}`,
			mockSetup: func(m *MockTopicUseCase) {
				m.On("CreateTopic", mock.Anything, mock.MatchedBy(func(topic *models.Topic) bool {
					return topic.Title == "Test Topic" && topic.Content == "Test Content"
				})).Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"message":"Topic created successfully!"`,
		},
		{
			name:           "Unauthorized",
			requestBody:    `{}`,
			mockSetup:      func(m *MockTopicUseCase) {},
			setupAuth:      func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `"error":"unauthorized"`,
		},
		{
			name: "Invalid Request Body",
			requestBody: `{
				"title": "Test Topic",
				"content": "Test Content"
			`,
			mockSetup: func(m *MockTopicUseCase) {},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"unexpected EOF"`,
		},
		{
			name: "Service Error",
			requestBody: `{
				"title": "Test Topic",
				"content": "Test Content"
			}`,
			mockSetup: func(m *MockTopicUseCase) {
				m.On("CreateTopic", mock.Anything, mock.AnythingOfType("*models.Topic")).
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
			mockUseCase := new(MockTopicUseCase)
			tt.mockSetup(mockUseCase)

			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewTopicHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.Use(tt.setupAuth)
			router.POST("/topics/", handler.CreateTopic)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/topics/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			if tt.mockSetup != nil {
				mockUseCase.AssertExpectations(t)
			}
		})
	}
}

func TestTopicHandler_GetAll(t *testing.T) {
	now := time.Now()
	testTopics := []*models.Topic{
		{Id: 1, Title: "Topic 1", Content: "Content 1", Username: "user1", CreatedAt: now},
		{Id: 2, Title: "Topic 2", Content: "Content 2", Username: "user2", CreatedAt: now},
	}
	tests := []struct {
		name           string
		mockSetup      func(*MockTopicUseCase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			mockSetup: func(m *MockTopicUseCase) {
				m.On("GetAllTopics", mock.Anything).
					Return(testTopics, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"data":[{"id":1,"title":"Topic 1"`,
		},
		{
			name: "Service Error",
			mockSetup: func(m *MockTopicUseCase) {
				m.On("GetAllTopics", mock.Anything).
					Return([]*models.Topic{}, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"service error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockTopicUseCase)
			tt.mockSetup(mockUseCase)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewTopicHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.GET("/topics/", handler.GetAll)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/topics/", nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestTopicHandler_GetTopic(t *testing.T) {
	now := time.Now()
	testTopic := &models.Topic{
		Id:        1,
		Title:     "Test Topic",
		Content:   "Test Content",
		Username:  "testuser",
		CreatedAt: now,
	}
	tests := []struct {
		name           string
		topicID        string
		mockSetup      func(*MockTopicUseCase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			topicID: "1",
			mockSetup: func(m *MockTopicUseCase) {
				m.On("GetTopic", mock.Anything, 1).
					Return(testTopic, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"id":1,"title":"Test Topic"`,
		},
		{
			name:           "Invalid ID",
			topicID:        "abc",
			mockSetup:      func(m *MockTopicUseCase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"strconv.Atoi: parsing \"abc\": invalid syntax"`,
		},
		{
			name:    "Service Error",
			topicID: "1",
			mockSetup: func(m *MockTopicUseCase) {
				m.On("GetTopic", mock.Anything, 1).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"service error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockTopicUseCase)
			tt.mockSetup(mockUseCase)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewTopicHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.GET("/topics/:id", handler.GetTopic)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/topics/"+tt.topicID, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			if tt.mockSetup != nil {
				mockUseCase.AssertExpectations(t)
			}
		})
	}
}

func TestTopicHandler_UpdateTopic(t *testing.T) {
	tests := []struct {
		name           string
		topicID        string
		requestBody    string
		mockSetup      func(*MockTopicUseCase)
		setupAuth      func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			topicID: "1",
			requestBody: `{
				"title": "Updated Topic",
				"content": "Updated Content"
			}`,
			mockSetup: func(m *MockTopicUseCase) {
				m.On("UpdateTopic", mock.Anything, mock.MatchedBy(func(topic *models.Topic) bool {
					return topic.Id == 1 && topic.Title == "Updated Topic"
				})).Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"Topic updated successfully!"`,
		},
		{
			name:        "Invalid ID",
			topicID:     "abc",
			requestBody: `{}`,
			mockSetup:   func(m *MockTopicUseCase) {},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"strconv.Atoi: parsing \"abc\": invalid syntax"`,
		},
		{
			name:    "Invalid Request Body",
			topicID: "1",
			requestBody: `{
				"title": "Updated Topic",
				"content": "Updated Content"
			`,
			mockSetup: func(m *MockTopicUseCase) {},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"unexpected EOF"`,
		},
		{
			name:    "Service Error",
			topicID: "1",
			requestBody: `{
				"title": "Updated Topic",
				"content": "Updated Content"
			}`,
			mockSetup: func(m *MockTopicUseCase) {
				m.On("UpdateTopic", mock.Anything, mock.AnythingOfType("*models.Topic")).
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
			mockUseCase := new(MockTopicUseCase)
			tt.mockSetup(mockUseCase)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewTopicHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.Use(tt.setupAuth)
			router.PUT("/topics/:id", handler.UpdateTopic)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/topics/"+tt.topicID, strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			if tt.mockSetup != nil {
				mockUseCase.AssertExpectations(t)
			} else {
				mockUseCase.AssertNotCalled(t, "UpdateTopic")
			}
		})
	}
}

func TestTopicHandler_DeleteTopic(t *testing.T) {
	tests := []struct {
		name           string
		topicID        string
		mockSetup      func(*MockTopicUseCase)
		setupAuth      func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			topicID: "1",
			mockSetup: func(m *MockTopicUseCase) {
				m.On("DeleteTopic", mock.Anything, 1).
					Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"Topic deleted successfully!"`,
		},
		{
			name:      "Invalid ID",
			topicID:   "abc",
			mockSetup: func(m *MockTopicUseCase) {},
			setupAuth: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"strconv.Atoi: parsing \"abc\": invalid syntax"`,
		},
		{
			name:    "Service Error",
			topicID: "1",
			mockSetup: func(m *MockTopicUseCase) {
				m.On("DeleteTopic", mock.Anything, 1).
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
			mockUseCase := new(MockTopicUseCase)
			tt.mockSetup(mockUseCase)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			handler := NewTopicHandler(mockUseCase, *logger)

			router := gin.New()
			router.Use(gin.Recovery())
			router.Use(tt.setupAuth)
			router.DELETE("/topics/:id", handler.DeleteTopic)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/topics/"+tt.topicID, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			if tt.mockSetup != nil {
				mockUseCase.AssertExpectations(t)
			} else {
				mockUseCase.AssertNotCalled(t, "DeleteTopic")
			}
		})
	}
}
