package postgres

import (
	"TopicService/internal/domain/models"
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCommentRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewCommentRepository(db)

	tests := []struct {
		name        string
		comment     *models.Comment
		mock        func()
		expectedID  int
		expectedErr error
	}{
		{
			name: "Success",
			comment: &models.Comment{
				TopicID:   1,
				Username:  "testuser",
				Content:   "test content",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mock: func() {
				mock.ExpectQuery("INSERT INTO comments").
					WithArgs(1, "testuser", "test content", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name: "Error",
			comment: &models.Comment{
				TopicID:   1,
				Username:  "testuser",
				Content:   "test content",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mock: func() {
				mock.ExpectQuery("INSERT INTO comments").
					WithArgs(1, "testuser", "test content", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedID:  0,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.Create(context.Background(), tt.comment)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.expectedID, tt.comment.Id)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestCommentRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewCommentRepository(db)

	tests := []struct {
		name        string
		comment     *models.Comment
		mock        func()
		expectedErr error
	}{
		{
			name: "Success",
			comment: &models.Comment{
				Id:      1,
				Content: "updated content",
			},
			mock: func() {
				mock.ExpectExec("UPDATE comments").
					WithArgs("updated content", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: nil,
		},
		{
			name: "NoRowsUpdated",
			comment: &models.Comment{
				Id:      1,
				Content: "updated content",
			},
			mock: func() {
				mock.ExpectExec("UPDATE comments").
					WithArgs("updated content", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedErr: errors.New("no rows updated"),
		},
		{
			name: "DatabaseError",
			comment: &models.Comment{
				Id:      1,
				Content: "updated content",
			},
			mock: func() {
				mock.ExpectExec("UPDATE comments").
					WithArgs("updated content", sqlmock.AnyArg(), 1).
					WillReturnError(errors.New("database error"))
			},
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.Update(context.Background(), tt.comment)

			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestCommentRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewCommentRepository(db)

	tests := []struct {
		name        string
		id          int
		mock        func()
		expectedErr error
	}{
		{
			name: "Success",
			id:   1,
			mock: func() {
				mock.ExpectExec("DELETE FROM comments").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: nil,
		},
		{
			name: "NoRowsDeleted",
			id:   1,
			mock: func() {
				mock.ExpectExec("DELETE FROM comments").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedErr: errors.New("no rows deleted"),
		},
		{
			name: "DatabaseError",
			id:   1,
			mock: func() {
				mock.ExpectExec("DELETE FROM comments").
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.Delete(context.Background(), tt.id)

			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestCommentRepository_FindById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewCommentRepository(db)

	now := time.Now()
	expectedComment := &models.Comment{
		Id:        1,
		TopicID:   1,
		Username:  "testuser",
		Content:   "test content",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name            string
		id              int
		mock            func()
		expectedComment *models.Comment
		expectedErr     error
	}{
		{
			name: "Success",
			id:   1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "topic_id", "username", "content", "created_at", "updated_at"}).
					AddRow(1, 1, "testuser", "test content", now, now)
				mock.ExpectQuery("SELECT \\* FROM comments WHERE id =").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedComment: expectedComment,
			expectedErr:     nil,
		},
		{
			name: "NotFound",
			id:   1,
			mock: func() {
				mock.ExpectQuery("SELECT \\* FROM comments WHERE id =").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			expectedComment: nil,
			expectedErr:     nil,
		},
		{
			name: "DatabaseError",
			id:   1,
			mock: func() {
				mock.ExpectQuery("SELECT \\* FROM comments WHERE id =").
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectedComment: nil,
			expectedErr:     errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			comment, err := repo.FindById(context.Background(), tt.id)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedComment != nil {
				assert.Equal(t, tt.expectedComment.Id, comment.Id)
				assert.Equal(t, tt.expectedComment.TopicID, comment.TopicID)
				assert.Equal(t, tt.expectedComment.Username, comment.Username)
				assert.Equal(t, tt.expectedComment.Content, comment.Content)
				// Can add more assertions for timestamps if needed
			} else {
				assert.Nil(t, comment)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestCommentRepository_FindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewCommentRepository(db)

	now := time.Now()
	expectedComments := []*models.Comment{
		{
			Id:        1,
			TopicID:   1,
			Username:  "testuser1",
			Content:   "test content 1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Id:        2,
			TopicID:   1,
			Username:  "testuser2",
			Content:   "test content 2",
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now.Add(-time.Hour),
		},
	}

	tests := []struct {
		name             string
		topicId          int
		mock             func()
		expectedComments []*models.Comment
		expectedErr      error
	}{
		{
			name:    "Success",
			topicId: 1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "topic_id", "username", "content", "created_at", "updated_at"}).
					AddRow(1, 1, "testuser1", "test content 1", now, now).
					AddRow(2, 1, "testuser2", "test content 2", now.Add(-time.Hour), now.Add(-time.Hour))
				mock.ExpectQuery("SELECT \\* FROM comments WHERE topic_id =").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedComments: expectedComments,
			expectedErr:      nil,
		},
		{
			name:    "NoComments",
			topicId: 1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "topic_id", "username", "content", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT \\* FROM comments WHERE topic_id =").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedComments: []*models.Comment{},
			expectedErr:      nil,
		},
		{
			name:    "DatabaseError",
			topicId: 1,
			mock: func() {
				mock.ExpectQuery("SELECT \\* FROM comments WHERE topic_id =").
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectedComments: nil,
			expectedErr:      errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			comments, err := repo.FindAll(context.Background(), tt.topicId)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedComments != nil {
				assert.Equal(t, len(tt.expectedComments), len(comments))
				for i := range tt.expectedComments {
					assert.Equal(t, tt.expectedComments[i].Id, comments[i].Id)
					assert.Equal(t, tt.expectedComments[i].TopicID, comments[i].TopicID)
					assert.Equal(t, tt.expectedComments[i].Username, comments[i].Username)
					assert.Equal(t, tt.expectedComments[i].Content, comments[i].Content)
				}
			} else {
				assert.Nil(t, comments)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
