package postgres

import (
	"AuthService/internal/domain"
	"AuthService/internal/domain/models"
	"AuthService/internal/domain/repositories"
	"context"
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

type UserRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewUserRepository(db *sql.DB, logger *zap.Logger) repositories.UserRepo {
	return &UserRepository{
		db:     db,
		logger: logger.With(zap.String("component", "user_repository")),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`

	r.logger.Debug("creating new user",
		zap.String("username", user.Username),
		zap.String("query", query))

	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password).Scan(&user.ID)
	if err != nil {
		r.logger.Error("failed to create user",
			zap.String("username", user.Username),
			zap.Error(err))
		return err
	}

	r.logger.Info("user created successfully",
		zap.Int("user_id", user.ID),
		zap.String("username", user.Username))
	return nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, password_hash FROM users WHERE username = $1`

	r.logger.Debug("searching user by username",
		zap.String("username", username),
		zap.String("query", query))

	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("user not found",
				zap.String("username", username))
			return nil, domain.UserNotFound
		}

		r.logger.Error("failed to find user by username",
			zap.String("username", username),
			zap.Error(err))
		return nil, err
	}

	r.logger.Debug("user found by username",
		zap.Int("user_id", user.ID),
		zap.String("username", user.Username))
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, password_hash FROM users WHERE id = $1`

	r.logger.Debug("searching user by ID",
		zap.Int("user_id", id),
		zap.String("query", query))

	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("user not found",
				zap.Int("user_id", id))
			return nil, domain.UserNotFound
		}

		r.logger.Error("failed to find user by ID",
			zap.Int("user_id", id),
			zap.Error(err))
		return nil, err
	}

	r.logger.Debug("user found by ID",
		zap.Int("user_id", user.ID),
		zap.String("username", user.Username))
	return &user, nil
}
