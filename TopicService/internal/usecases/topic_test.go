package usecases_test

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"TopicService/internal/domain/models"
	"TopicService/internal/usecases"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTopicRepo struct {
	mock.Mock
}

func (m *MockTopicRepo) Create(ctx context.Context, topic *models.Topic) error {
	args := m.Called(ctx, topic)
	return args.Error(0)
}

func (m *MockTopicRepo) FindById(ctx context.Context, id int) (*models.Topic, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Topic), args.Error(1)
}

func (m *MockTopicRepo) FindAll(ctx context.Context) ([]*models.Topic, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Topic), args.Error(1)
}

func (m *MockTopicRepo) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTopicRepo) Update(ctx context.Context, topic *models.Topic) error {
	args := m.Called(ctx, topic)
	return args.Error(0)
}

func TestTopicService_CreateTopic(t *testing.T) {
	tests := []struct {
		name        string
		topic       *models.Topic
		repoError   error
		expectedErr error
	}{
		{
			name: "successful creation",
			topic: &models.Topic{
				Title:    "Test Topic",
				Content:  "Test Content",
				Username: "testuser",
			},
			repoError:   nil,
			expectedErr: nil,
		},
		{
			name: "repository error",
			topic: &models.Topic{
				Title:    "Test Topic",
				Content:  "Test Content",
				Username: "testuser",
			},
			repoError:   errors.New("database error"),
			expectedErr: errors.New("failed to create topic"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTopicRepo)
			mockRepo.On("Create", mock.Anything, tt.topic).Return(tt.repoError)

			service := usecases.NewTopicUseCase(mockRepo, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
			err := service.CreateTopic(context.Background(), tt.topic)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTopicService_GetTopic(t *testing.T) {
	now := time.Now()
	testTopic := &models.Topic{
		Id:        1,
		Title:     "Test Topic",
		Content:   "Test Content",
		Username:  "testuser",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name        string
		id          int
		repoResult  *models.Topic
		repoError   error
		expected    *models.Topic
		expectedErr error
	}{
		{
			name:        "successful get",
			id:          1,
			repoResult:  testTopic,
			repoError:   nil,
			expected:    testTopic,
			expectedErr: nil,
		},
		{
			name:        "not found",
			id:          2,
			repoResult:  nil,
			repoError:   sql.ErrNoRows,
			expected:    nil,
			expectedErr: errors.New("topic not found"),
		},
		{
			name:        "database error",
			id:          3,
			repoResult:  nil,
			repoError:   errors.New("database error"),
			expected:    nil,
			expectedErr: errors.New("failed to get topic"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTopicRepo)
			mockRepo.On("FindById", mock.Anything, tt.id).Return(tt.repoResult, tt.repoError)

			service := usecases.NewTopicUseCase(mockRepo, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
			result, err := service.GetTopic(context.Background(), tt.id)

			assert.Equal(t, tt.expected, result)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTopicService_GetAllTopics(t *testing.T) {
	now := time.Now()
	testTopics := []*models.Topic{
		{
			Id:        1,
			Title:     "Topic 1",
			Content:   "Content 1",
			Username:  "user1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Id:        2,
			Title:     "Topic 2",
			Content:   "Content 2",
			Username:  "user2",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	tests := []struct {
		name        string
		repoResult  []*models.Topic
		repoError   error
		expected    []*models.Topic
		expectedErr error
	}{
		{
			name:        "successful get all",
			repoResult:  testTopics,
			repoError:   nil,
			expected:    testTopics,
			expectedErr: nil,
		},
		{
			name:        "empty list",
			repoResult:  []*models.Topic{},
			repoError:   nil,
			expected:    []*models.Topic{},
			expectedErr: nil,
		},
		{
			name:        "repository error",
			repoResult:  nil,
			repoError:   errors.New("database error"),
			expected:    nil,
			expectedErr: errors.New("failed to get topics"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTopicRepo)
			mockRepo.On("FindAll", mock.Anything).Return(tt.repoResult, tt.repoError)

			service := usecases.NewTopicUseCase(mockRepo, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
			result, err := service.GetAllTopics(context.Background())

			assert.Equal(t, tt.expected, result)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTopicService_DeleteTopic(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		repoError   error
		expectedErr error
	}{
		{
			name:        "successful delete",
			id:          1,
			repoError:   nil,
			expectedErr: nil,
		},
		{
			name:        "not found",
			id:          2,
			repoError:   errors.New("no rows deleted"),
			expectedErr: errors.New("failed to delete topic"),
		},
		{
			name:        "repository error",
			id:          3,
			repoError:   errors.New("database error"),
			expectedErr: errors.New("failed to delete topic"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTopicRepo)
			mockRepo.On("Delete", mock.Anything, tt.id).Return(tt.repoError)

			service := usecases.NewTopicUseCase(mockRepo, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
			err := service.DeleteTopic(context.Background(), tt.id)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTopicService_UpdateTopic(t *testing.T) {
	now := time.Now()
	testTopic := &models.Topic{
		Id:        1,
		Title:     "Updated Title",
		Content:   "Updated Content",
		Username:  "testuser",
		UpdatedAt: now,
	}

	tests := []struct {
		name        string
		topic       *models.Topic
		repoError   error
		expectedErr error
	}{
		{
			name:        "successful update",
			topic:       testTopic,
			repoError:   nil,
			expectedErr: nil,
		},
		{
			name: "not found",
			topic: &models.Topic{
				Id:        2,
				Title:     "Non-existent",
				Content:   "Content",
				Username:  "user",
				UpdatedAt: now,
			},
			repoError:   errors.New("no rows updated"),
			expectedErr: errors.New("failed to update topic"),
		},
		{
			name:        "repository error",
			topic:       testTopic,
			repoError:   errors.New("database error"),
			expectedErr: errors.New("failed to update topic"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTopicRepo)
			mockRepo.On("Update", mock.Anything, tt.topic).Return(tt.repoError)

			service := usecases.NewTopicUseCase(mockRepo, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
			err := service.UpdateTopic(context.Background(), tt.topic)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
