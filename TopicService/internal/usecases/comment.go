package usecases

import (
	"TopicService/internal/domain/models"
	"TopicService/internal/interfaces/api/persistence/postgres"
	"context"
	"errors"
	"log/slog"
)

type CommentService struct {
	repo   postgres.CommentRepo
	logger slog.Logger
}

func NewCommentUseCase(repo postgres.CommentRepo, logger slog.Logger) CommentUseCasesInterface {
	return &CommentService{
		repo:   repo,
		logger: logger,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, comment *models.Comment) error {
	err := s.repo.Create(ctx, comment)
	if err != nil {
		s.logger.Error("Ошибка создания комментария",
			"error", err,
			"topicID", comment.TopicID,
			"username", comment.Username)
		return errors.New("failed to create comment")
	}

	s.logger.Info("Комментарий успешно создан",
		"id", comment.Id,
		"topicID", comment.TopicID)
	return nil
}

func (s *CommentService) GetComment(ctx context.Context, id int) (*models.Comment, error) {

	comment, err := s.repo.FindById(ctx, id)
	if err != nil {
		if errors.As(err, &models.ErrNotFound{}) {
			s.logger.Warn("Комментарий не найден", "id", id)
			return nil, err
		}
		s.logger.Error("Ошибка получения темы",
			"error", err,
			"id", id)
		return nil, errors.New("failed to get comment")
	}

	s.logger.Info("Комментарий успешно найден",
		"id", comment.Id)
	return comment, nil
}

func (s *CommentService) GetAllComments(ctx context.Context, topicId int) ([]*models.Comment, error) {

	comments, err := s.repo.FindAll(ctx, topicId)
	if err != nil {
		s.logger.Error("Ошибка получения комментариев",
			"error", err,
			"topicID", topicId)
		return nil, errors.New("failed to get comments")
	}

	s.logger.Info("Комментарии успешно получены",
		"topicID", topicId,
		"count", len(comments))
	return comments, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, id int) error {

	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Ошибка удаления комментария",
			"error", err,
			"id", id)
		return errors.New("failed to delete comment")
	}

	s.logger.Info("Комментарий успешно удален", "id", id)
	return nil
}

func (s *CommentService) UpdateComment(ctx context.Context, comment *models.Comment) error {

	err := s.repo.Update(ctx, comment)
	if err != nil {
		s.logger.Error("Ошибка обновления комментария",
			"error", err,
			"id", comment.Id)
		return errors.New("failed to update comment")
	}

	s.logger.Info("Комментарий успешно обновлен",
		"id", comment.Id,
		"topicID", comment.TopicID)
	return nil
}
