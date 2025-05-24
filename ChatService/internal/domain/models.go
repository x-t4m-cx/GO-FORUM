package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ChatMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username"`
	Message   string             `bson:"message"`
	CreatedAt time.Time          `bson:"created_at"`
	ExpiresAt time.Time          `bson:"expires_at"`
}
