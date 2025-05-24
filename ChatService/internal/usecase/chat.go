package usecase

import (
	"ChatService/internal/domain"
	"time"
)

type ChatUsecase interface {
	ProcessMessage(username, text string) (domain.ChatMessage, error)
	GetRecentMessages(limit int) ([]domain.ChatMessage, error)
	CleanupExpiredMessages() error
}

type chatUsecase struct {
	repo            domain.ChatRepository
	messageLifetime time.Duration
}

func NewChatUsecase(repo domain.ChatRepository, lifetime time.Duration) ChatUsecase {
	return &chatUsecase{
		repo:            repo,
		messageLifetime: lifetime,
	}
}

func (c *chatUsecase) ProcessMessage(username, text string) (domain.ChatMessage, error) {
	message := domain.ChatMessage{
		Username:  username,
		Message:   text,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(c.messageLifetime),
	}

	err := c.repo.SaveMessage(&message)
	if err != nil {
		return domain.ChatMessage{}, err
	}

	return message, nil
}

func (c *chatUsecase) GetRecentMessages(limit int) ([]domain.ChatMessage, error) {
	return c.repo.GetRecentMessages(limit)
}

func (c *chatUsecase) CleanupExpiredMessages() error {
	return c.repo.CleanupExpiredMessages()
}
