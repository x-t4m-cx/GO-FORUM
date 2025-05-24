package config

import (
	"github.com/joho/godotenv"
	"os"
	"time"
)

type Config struct {
	AppEnv        string
	ServerPort    string
	GRPCPort      string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPass        string
	DBName        string
	DBSSLMode     string
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	accessTTL, err := time.ParseDuration(getEnv("AccessTTL", "15m"))
	if err != nil {
		return nil, err
	}

	refreshTTL, err := time.ParseDuration(getEnv("RefreshTTL", "720h"))
	if err != nil {
		return nil, err
	}
	return &Config{
		AppEnv:        getEnv("AppEnv", "development"),
		ServerPort:    getEnv("ServerPort", "8081"),
		GRPCPort:      getEnv("GRPCPort", ""),
		DBHost:        getEnv("DBHost", "localhost"),
		DBPort:        getEnv("DBPort", "5432"),
		DBUser:        getEnv("DBUser", "postgres"),
		DBPass:        getEnv("DBPass", ""),
		DBName:        getEnv("DBName", "userDB"),
		DBSSLMode:     getEnv("DBSSLMode", ""),
		AccessSecret:  getEnv("AccessSecret", ""),
		RefreshSecret: getEnv("RefreshSecret", ""),
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
	}, nil
}
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
