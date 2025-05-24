//go:build integration
// +build integration

package postgres

import (
	"TopicService/internal/domain/models"
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq" // Драйвер PostgreSQL
)

// Теги для интеграционных тестов
var testDB *sql.DB

func TestMain(m *testing.M) {
	// Настройка БД перед тестами
	setupDB()

	// Запуск тестов
	code := m.Run()

	// Очистка после тестов
	teardownDB()

	os.Exit(code)
}

func setupDB() {
	var err error
	connStr := "user=postgres dbname=test_db password=postgres sslmode=disable"
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS topics (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			username TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func teardownDB() {
	_, err := testDB.Exec("DROP TABLE IF EXISTS topics")
	if err != nil {
		log.Fatalf("Failed to drop table: %v", err)
	}
	testDB.Close()
}

func TestTopicRepository_Create(t *testing.T) {
	repo := NewTopicRepository(testDB)
	ctx := context.Background()

	topic := &models.Topic{
		Title:     "Test Topic",
		Content:   "This is a test topic",
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, topic)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if topic.Id <= 0 {
		t.Error("Expected positive ID, got:", topic.Id)
	}

	foundTopic, err := repo.FindById(ctx, topic.Id)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}
	if foundTopic == nil {
		t.Fatal("Topic not found")
	}
	if foundTopic.Title != topic.Title {
		t.Errorf("Expected title '%s', got '%s'", topic.Title, foundTopic.Title)
	}
}

func TestTopicRepository_Update(t *testing.T) {
	repo := NewTopicRepository(testDB)
	ctx := context.Background()

	topic := &models.Topic{
		Title:     "Old Title",
		Content:   "Old Content",
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, topic)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	newTitle := "Updated Title"
	newContent := "Updated Content"
	topic.Title = newTitle
	topic.Content = newContent

	err = repo.Update(ctx, topic)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updatedTopic, err := repo.FindById(ctx, topic.Id)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}
	if updatedTopic.Title != newTitle {
		t.Errorf("Expected title '%s', got '%s'", newTitle, updatedTopic.Title)
	}
	if updatedTopic.Content != newContent {
		t.Errorf("Expected content '%s', got '%s'", newContent, updatedTopic.Content)
	}
}

func TestTopicRepository_Delete(t *testing.T) {
	repo := NewTopicRepository(testDB)
	ctx := context.Background()

	topic := &models.Topic{
		Title:     "Topic to Delete",
		Content:   "This will be deleted",
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, topic)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = repo.Delete(ctx, topic.Id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	deletedTopic, err := repo.FindById(ctx, topic.Id)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Expected sql.ErrNoRows, got: %v", err)
	}
	if deletedTopic != nil {
		t.Error("Expected topic to be deleted, but it was found")
	}
}

func TestTopicRepository_FindById_NotFound(t *testing.T) {
	repo := NewTopicRepository(testDB)
	ctx := context.Background()

	_, err := repo.FindById(ctx, 9999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Expected sql.ErrNoRows, got: %v", err)
	}
}

func TestTopicRepository_FindAll(t *testing.T) {
	repo := NewTopicRepository(testDB)
	ctx := context.Background()

	_, err := testDB.Exec("TRUNCATE TABLE topics")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	topics := []*models.Topic{
		{
			Title:     "Topic 1",
			Content:   "Content 1",
			Username:  "user1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Title:     "Topic 2",
			Content:   "Content 2",
			Username:  "user2",
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now(),
		},
	}

	for _, topic := range topics {
		err := repo.Create(ctx, topic)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	foundTopics, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(foundTopics) != 2 {
		t.Fatalf("Expected 2 topics, got %d", len(foundTopics))
	}

	if foundTopics[0].Title != "Topic 1" {
		t.Errorf("Expected first topic to be 'Topic 1', got '%s'", foundTopics[0].Title)
	}
}
