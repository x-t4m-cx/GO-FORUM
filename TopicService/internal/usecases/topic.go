package usecases

import (
	"TopicService/internal/domain/models"
	"TopicService/internal/interfaces/api/persistence/postgres"
	"context"
	"errors"
	"log/slog"
)

type TopicService struct {
	repo   postgres.TopicRepo
	logger slog.Logger
}

func NewTopicUseCase(repo postgres.TopicRepo, logger slog.Logger) TopicUseCasesInterface {
	return &TopicService{
		repo:   repo,
		logger: logger,
	}
}

func (s *TopicService) CreateTopic(ctx context.Context, topic *models.Topic) error {

	err := s.repo.Create(ctx, topic)
	if err != nil {
		s.logger.Error("Ошибка создания темы",
			"error", err,
			"topic", topic.Title)
		return errors.New("failed to create topic")
	}

	s.logger.Info("Тема успешно создана",
		"id", topic.Id,
		"title", topic.Title)
	return nil
}

func (s *TopicService) GetTopic(ctx context.Context, id int) (*models.Topic, error) {

	topic, err := s.repo.FindById(ctx, id)
	if err != nil {
		if errors.As(err, &models.ErrNotFound{}) {
			s.logger.Warn("Тема не найдена", "id", id)
			return nil, err
		}

		s.logger.Error("Ошибка получения темы",
			"error", err,
			"id", id)
		return nil, errors.New("failed to get topic")
	}

	s.logger.Info("Тема успешно найдена",
		"id", topic.Id,
		"title", topic.Title)
	return topic, nil
}

func (s *TopicService) GetAllTopics(ctx context.Context) ([]*models.Topic, error) {

	topics, err := s.repo.FindAll(ctx)
	if err != nil {
		s.logger.Error("Ошибка получения тем", "error", err)
		return nil, errors.New("failed to get topics")
	}

	s.logger.Info("Темы успешно найдены", "count", len(topics))
	return topics, nil
}

func (s *TopicService) DeleteTopic(ctx context.Context, id int) error {

	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Ошибка удаления темы",
			"error", err,
			"id", id)
		return errors.New("failed to delete topic")
	}

	s.logger.Info("Тема успешно удалена", "id", id)
	return nil
}

func (s *TopicService) UpdateTopic(ctx context.Context, topic *models.Topic) error {

	err := s.repo.Update(ctx, topic)
	if err != nil {
		s.logger.Error("Ошибка обновления темы",
			"error", err,
			"id", topic.Id)
		return errors.New("failed to update topic")
	}

	s.logger.Info("тема успешно обновлена",
		"id", topic.Id,
		"title", topic.Title)
	return nil
}
