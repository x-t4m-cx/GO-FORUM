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

// MockCommentRepo is a mock implementation of CommentRepo for testing
type MockCommentRepo struct {
	mock.Mock
}

func (m *MockCommentRepo) Create(ctx context.Context, comment *models.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepo) FindById(ctx context.Context, id int) (*models.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Comment), args.Error(1)
}

func (m *MockCommentRepo) FindAll(ctx context.Context, topicId int) ([]*models.Comment, error) {
	args := m.Called(ctx, topicId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Comment), args.Error(1)
}

func (m *MockCommentRepo) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepo) Update(ctx context.Context, comment *models.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func TestCommentUseCase_CreateComment(t *testing.T) {
	tests := []struct {
		name        string
		comment     *models.Comment
		repoError   error
		expectedErr error
	}{
		{
			name: "successful creation",
			comment: &models.Comment{
				TopicID:   1,
				Username:  "testuser",
				Content:   "Test content",
				CreatedAt: time.Now(),
			},
			repoError:   nil,
			expectedErr: nil,
		},
		{
			name: "repository error on create",
			comment: &models.Comment{
				TopicID:   1,
				Username:  "testuser",
				Content:   "Test content",
				CreatedAt: time.Now(),
			},
			repoError:   errors.New("database error"),
			expectedErr: errors.New("failed to create comment"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCommentRepo)
			mockRepo.On("Create", mock.Anything, tt.comment).Return(tt.repoError)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			service := usecases.NewCommentUseCase(mockRepo, *logger)
			err := service.CreateComment(context.Background(), tt.comment)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCommentUseCase_GetComment(t *testing.T) {
	now := time.Now()
	testComment := &models.Comment{
		Id:        1,
		TopicID:   1,
		Username:  "testuser",
		Content:   "Test content",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name        string
		id          int
		repoResult  *models.Comment
		repoError   error
		expected    *models.Comment
		expectedErr error
	}{
		{
			name:        "successful get",
			id:          1,
			repoResult:  testComment,
			repoError:   nil,
			expected:    testComment,
			expectedErr: nil,
		},
		{
			name:        "not found",
			id:          2,
			repoResult:  nil,
			repoError:   sql.ErrNoRows,
			expected:    nil,
			expectedErr: sql.ErrNoRows,
		},
		{
			name:        "repository error",
			id:          3,
			repoResult:  nil,
			repoError:   errors.New("database error"),
			expected:    nil,
			expectedErr: errors.New("failed to get comment"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCommentRepo)
			mockRepo.On("FindById", mock.Anything, tt.id).Return(tt.repoResult, tt.repoError)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			service := usecases.NewCommentUseCase(mockRepo, *logger)
			result, err := service.GetComment(context.Background(), tt.id)

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

func TestCommentUseCase_GetAllComments(t *testing.T) {
	now := time.Now()
	testComments := []*models.Comment{
		{
			Id:        1,
			TopicID:   1,
			Username:  "user1",
			Content:   "Content 1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Id:        2,
			TopicID:   1,
			Username:  "user2",
			Content:   "Content 2",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	tests := []struct {
		name        string
		topicId     int
		repoResult  []*models.Comment
		repoError   error
		expected    []*models.Comment
		expectedErr error
	}{
		{
			name:        "successful get all",
			topicId:     1,
			repoResult:  testComments,
			repoError:   nil,
			expected:    testComments,
			expectedErr: nil,
		},
		{
			name:        "empty list",
			topicId:     2,
			repoResult:  []*models.Comment{},
			repoError:   nil,
			expected:    []*models.Comment{},
			expectedErr: nil,
		},
		{
			name:        "repository error",
			topicId:     3,
			repoResult:  nil,
			repoError:   errors.New("database error"),
			expected:    nil,
			expectedErr: errors.New("failed to get comments"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCommentRepo)
			mockRepo.On("FindAll", mock.Anything, tt.topicId).Return(tt.repoResult, tt.repoError)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
			service := usecases.NewCommentUseCase(mockRepo, *logger)
			result, err := service.GetAllComments(context.Background(), tt.topicId)

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

func TestCommentUseCase_UpdateComment(t *testing.T) {
	now := time.Now()
	testComment := &models.Comment{
		Id:        1,
		Content:   "Updated content",
		UpdatedAt: now,
	}

	tests := []struct {
		name        string
		comment     *models.Comment
		repoError   error
		expectedErr error
	}{
		{
			name:        "successful update",
			comment:     testComment,
			repoError:   nil,
			expectedErr: nil,
		},
		{
			name:        "not found",
			comment:     testComment,
			repoError:   errors.New("no rows updated"),
			expectedErr: errors.New("failed to update comment"),
		},
		{
			name:        "repository error",
			comment:     testComment,
			repoError:   errors.New("database error"),
			expectedErr: errors.New("failed to update comment"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCommentRepo)
			mockRepo.On("Update", mock.Anything, tt.comment).Return(tt.repoError)

			service := usecases.NewCommentUseCase(mockRepo, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
			err := service.UpdateComment(context.Background(), tt.comment)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCommentUseCase_DeleteComment(t *testing.T) {
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
			expectedErr: errors.New("failed to delete comment"),
		},
		{
			name:        "repository error",
			id:          3,
			repoError:   errors.New("database error"),
			expectedErr: errors.New("failed to delete comment"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCommentRepo)
			mockRepo.On("Delete", mock.Anything, tt.id).Return(tt.repoError)

			service := usecases.NewCommentUseCase(mockRepo, *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
			err := service.DeleteComment(context.Background(), tt.id)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
