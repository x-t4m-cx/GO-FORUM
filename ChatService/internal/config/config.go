package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Config struct {
	ServerAddress   string
	MongoDBURI      string
	DatabaseName    string
	CollectionName  string
	MessageLifetime time.Duration
}

func LoadConfig() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return &Config{
		ServerAddress:   getEnv("PORT", "8090"),
		MongoDBURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DatabaseName:    getEnv("DATABASE_NAME", "chat_db"),
		CollectionName:  getEnv("COLLECTION_NAME", "messages"),
		MessageLifetime: 1 * time.Minute,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
