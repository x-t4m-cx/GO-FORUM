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

func TestTopicRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewTopicRepository(db)

	tests := []struct {
		name        string
		topic       *models.Topic
		mock        func()
		expectedID  int
		expectedErr error
	}{
		{
			"Success",
			&models.Topic{
				Title:     "Test Topic",
				Content:   "Test Content",
				Username:  "testuser",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			func() {
				mock.ExpectQuery("INSERT INTO topics").
					WithArgs("Test Topic", "Test Content", "testuser", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			1,
			nil,
		},
		{
			name: "Error",
			topic: &models.Topic{
				Title:     "Test Topic",
				Content:   "Test Content",
				Username:  "testuser",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mock: func() {
				mock.ExpectQuery("INSERT INTO topics").
					WithArgs("Test Topic", "Test Content", "testuser", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedID:  0,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.Create(context.Background(), tt.topic)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.expectedID, tt.topic.Id)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestTopicRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewTopicRepository(db)

	tests := []struct {
		name        string
		topic       *models.Topic
		mock        func()
		expectedErr error
	}{
		{
			name: "Success",
			topic: &models.Topic{
				Id:      1,
				Title:   "Updated Title",
				Content: "Updated Content",
			},
			mock: func() {
				mock.ExpectExec("UPDATE topics").
					WithArgs("Updated Title", "Updated Content", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: nil,
		},
		{
			name: "DatabaseError",
			topic: &models.Topic{
				Id:      1,
				Title:   "Updated Title",
				Content: "Updated Content",
			},
			mock: func() {
				mock.ExpectExec("UPDATE topics").
					WithArgs("Updated Title", "Updated Content", sqlmock.AnyArg(), 1).
					WillReturnError(errors.New("database error"))
			},
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.Update(context.Background(), tt.topic)

			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestTopicRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewTopicRepository(db)

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
				mock.ExpectExec("DELETE FROM topics").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: nil,
		},
		{
			name: "DatabaseError",
			id:   1,
			mock: func() {
				mock.ExpectExec("DELETE FROM topics").
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

func TestTopicRepository_FindById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewTopicRepository(db)

	now := time.Now()
	expectedTopic := &models.Topic{
		Id:        1,
		Title:     "Test Topic",
		Content:   "Test Content",
		Username:  "testuser",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name          string
		id            int
		mock          func()
		expectedTopic *models.Topic
		expectedErr   error
	}{
		{
			name: "Success",
			id:   1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "content", "username", "created_at", "updated_at"}).
					AddRow(1, "Test Topic", "Test Content", "testuser", now, now)
				mock.ExpectQuery("SELECT \\* FROM topics WHERE id =").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedTopic: expectedTopic,
			expectedErr:   nil,
		},
		{
			name: "NotFound",
			id:   1,
			mock: func() {
				mock.ExpectQuery("SELECT \\* FROM topics WHERE id =").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			expectedTopic: nil,
			expectedErr:   &models.ErrNotFound{},
		},
		{
			name: "DatabaseError",
			id:   1,
			mock: func() {
				mock.ExpectQuery("SELECT \\* FROM topics WHERE id =").
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectedTopic: nil,
			expectedErr:   errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			topic, err := repo.FindById(context.Background(), tt.id)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedTopic != nil {
				assert.Equal(t, tt.expectedTopic.Id, topic.Id)
				assert.Equal(t, tt.expectedTopic.Title, topic.Title)
				assert.Equal(t, tt.expectedTopic.Content, topic.Content)
				assert.Equal(t, tt.expectedTopic.Username, topic.Username)
			} else {
				assert.Nil(t, topic)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestTopicRepository_FindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewTopicRepository(db)

	now := time.Now()
	expectedTopics := []*models.Topic{
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
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now.Add(-time.Hour),
		},
	}

	tests := []struct {
		name           string
		mock           func()
		expectedTopics []*models.Topic
		expectedErr    error
	}{
		{
			name: "Success",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "content", "username", "created_at", "updated_at"}).
					AddRow(1, "Topic 1", "Content 1", "user1", now, now).
					AddRow(2, "Topic 2", "Content 2", "user2", now.Add(-time.Hour), now.Add(-time.Hour))
				mock.ExpectQuery("SELECT \\* FROM topics ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedTopics: expectedTopics,
			expectedErr:    nil,
		},
		{
			name: "NoTopics",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "content", "username", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT \\* FROM topics ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedTopics: []*models.Topic{},
			expectedErr:    nil,
		},
		{
			name: "DatabaseError",
			mock: func() {
				mock.ExpectQuery("SELECT \\* FROM topics ORDER BY created_at DESC").
					WillReturnError(errors.New("database error"))
			},
			expectedTopics: nil,
			expectedErr:    errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			topics, err := repo.FindAll(context.Background())

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedTopics != nil {
				assert.Equal(t, len(tt.expectedTopics), len(topics))
				for i := range tt.expectedTopics {
					assert.Equal(t, tt.expectedTopics[i].Id, topics[i].Id)
					assert.Equal(t, tt.expectedTopics[i].Title, topics[i].Title)
					assert.Equal(t, tt.expectedTopics[i].Content, topics[i].Content)
					assert.Equal(t, tt.expectedTopics[i].Username, topics[i].Username)
				}
			} else {
				assert.Nil(t, topics)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
