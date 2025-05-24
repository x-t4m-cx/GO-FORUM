package pg

import (
	"AuthService/internal/config"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

// DBConfig содержит конфигурацию подключения к БД
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDB создает новое подключение к PostgreSQL
func NewDB(cfg *config.Config) (*sql.DB, error) {
	dbConfig := DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPass,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.DBName,
		dbConfig.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return db, nil
}

// CloseDB закрывает подключение к БД
func CloseDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}
}
