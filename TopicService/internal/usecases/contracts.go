package usecases

import (
	"TopicService/internal/domain/models"
	"context"
	"net/http"
)

type CommentUseCasesInterface interface {
	CreateComment(ctx context.Context, topic *models.Comment) error
	GetComment(ctx context.Context, id int) (*models.Comment, error)
	GetAllComments(ctx context.Context, topicId int) ([]*models.Comment, error)
	DeleteComment(ctx context.Context, id int) error
	UpdateComment(ctx context.Context, topic *models.Comment) error
}
type TopicUseCasesInterface interface {
	CreateTopic(ctx context.Context, topic *models.Topic) error
	GetTopic(ctx context.Context, id int) (*models.Topic, error)
	GetAllTopics(ctx context.Context) ([]*models.Topic, error)
	DeleteTopic(ctx context.Context, id int) error
	UpdateTopic(ctx context.Context, topic *models.Topic) error
}
type GRPCClientInterface interface {
	Login(ctx context.Context, username, password string) (*http.Response, error)
	Logout(ctx context.Context) (*http.Response, error)
	Register(ctx context.Context, username, password string) (*http.Response, error)
	Refresh(ctx context.Context, refreshToken string) (*http.Response, error)
	VerifyToken(ctx context.Context, token string) (string, error)
	Close() error
}
