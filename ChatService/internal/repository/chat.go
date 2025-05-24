package repository

import (
	"ChatService/internal/config"
	"ChatService/internal/domain"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type mongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoRepository(cfg *config.Config) (domain.ChatRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDBURI))
	if err != nil {
		return nil, err
	}

	collection := client.Database(cfg.DatabaseName).Collection(cfg.CollectionName)

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"expires_at": 1},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, err
	}

	return &mongoRepository{
		client:     client,
		collection: collection,
	}, nil
}

func (m *mongoRepository) SaveMessage(message *domain.ChatMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.collection.InsertOne(ctx, message)
	return err
}

func (m *mongoRepository) GetRecentMessages(limit int) ([]domain.ChatMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"expires_at": bson.M{"$gt": time.Now()}}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(int64(limit))

	cursor, err := m.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []domain.ChatMessage
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (m *mongoRepository) CleanupExpiredMessages() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.collection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": time.Now()}})
	return err
}
