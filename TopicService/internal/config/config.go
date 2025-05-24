package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	AppEnv      string
	ServerPort  string
	AuthService string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	LogLevel  string `mapstructure:"LOG_LEVEL"`
	LogFormat string `mapstructure:"LOG_FORMAT"`
	LogOutput string `mapstructure:"LOG_OUTPUT"`
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Не удалось загрузить .env файл: %v", err)
	}
	return &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		AuthService: getEnv("AUTH_SERVICE", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", ""),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", ""),
		DBSSLMode:   getEnv("DB_SSL_MODE", "disable"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogFormat:   getEnv("LOG_FORMAT", "text"),
		LogOutput:   getEnv("LOG_OUTPUT", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
