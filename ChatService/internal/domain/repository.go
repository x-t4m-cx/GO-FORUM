package domain

type ChatRepository interface {
	SaveMessage(message *ChatMessage) error
	GetRecentMessages(limit int) ([]ChatMessage, error)
	CleanupExpiredMessages() error
}
