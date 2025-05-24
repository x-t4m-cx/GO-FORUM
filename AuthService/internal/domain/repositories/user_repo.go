package repositories

import (
	"AuthService/internal/domain/models"
	"context"
)

type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByID(ctx context.Context, id int) (*models.User, error)
}
