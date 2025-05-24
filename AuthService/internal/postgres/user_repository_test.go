package postgres_test

import (
	"AuthService/internal/domain"
	"AuthService/internal/domain/models"
	"AuthService/internal/postgres"

	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		user        *models.User
		mock        func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "Success",
			user: &models.User{
				Username: "testuser",
				Password: "hashedpassword",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", "hashedpassword").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedErr: nil,
		},
		{
			name: "Duplicate Username",
			user: &models.User{
				Username: "existinguser",
				Password: "hashedpassword",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("existinguser", "hashedpassword").
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
			},
			expectedErr: errors.New("duplicate key value violates unique constraint"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.mock(mock)

			repo := postgres.NewUserRepository(db, nil)
			err = repo.Create(context.Background(), tt.user)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr == nil {
				assert.Equal(t, 1, tt.user.ID)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		mock        func(mock sqlmock.Sqlmock)
		expected    *models.User
		expectedErr error
	}{
		{
			name:     "Success",
			username: "testuser",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow(1, "testuser", "hashedpassword")
				mock.ExpectQuery("SELECT id, username, password_hash FROM users WHERE username = \\$1").
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			expected: &models.User{
				ID:       1,
				Username: "testuser",
				Password: "hashedpassword",
			},
			expectedErr: nil,
		},
		{
			name:     "User Not Found",
			username: "nonexistent",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password_hash FROM users WHERE username = \\$1").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectedErr: domain.UserNotFound,
		},
		{
			name:     "Database Error",
			username: "testuser",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password_hash FROM users WHERE username = \\$1").
					WithArgs("testuser").
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.mock(mock)

			repo := postgres.NewUserRepository(db, nil)
			user, err := repo.FindByUsername(context.Background(), tt.username)

			assert.Equal(t, tt.expected, user)
			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		mock        func(mock sqlmock.Sqlmock)
		expected    *models.User
		expectedErr error
	}{
		{
			name: "Success",
			id:   1,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow(1, "testuser", "hashedpassword")
				mock.ExpectQuery("SELECT id, username, password_hash FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expected: &models.User{
				ID:       1,
				Username: "testuser",
				Password: "hashedpassword",
			},
			expectedErr: nil,
		},
		{
			name: "User Not Found",
			id:   999,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password_hash FROM users WHERE id = \\$1").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectedErr: domain.UserNotFound,
		},
		{
			name: "Database Error",
			id:   1,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password_hash FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.mock(mock)

			repo := postgres.NewUserRepository(db, nil)
			user, err := repo.FindByID(context.Background(), tt.id)

			assert.Equal(t, tt.expected, user)
			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
