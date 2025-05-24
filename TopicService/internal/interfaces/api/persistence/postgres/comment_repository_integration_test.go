//go:build integration
// +build integration

package postgres

import (
	"TopicService/internal/domain/models"
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	setupDB()
	code := m.Run()
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
		CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			topic_id INTEGER NOT NULL,
			username TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func teardownDB() {
	_, err := testDB.Exec("DROP TABLE IF EXISTS comments")
	if err != nil {
		log.Fatalf("Failed to drop table: %v", err)
	}
	testDB.Close()
}

func TestCommentRepository_Create(t *testing.T) {
	repo := NewCommentRepository(testDB)
	ctx := context.Background()

	comment := &models.Comment{
		TopicID:   1,
		Username:  "testuser",
		Content:   "Test comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, comment)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if comment.Id <= 0 {
		t.Error("Expected positive ID, got:", comment.Id)
	}

	// Verify creation
	found, err := repo.FindById(ctx, comment.Id)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}
	if found == nil {
		t.Fatal("Comment not found")
	}
	if found.Content != comment.Content {
		t.Errorf("Expected content '%s', got '%s'", comment.Content, found.Content)
	}
}

func TestCommentRepository_Update(t *testing.T) {
	repo := NewCommentRepository(testDB)
	ctx := context.Background()

	// Create test comment
	comment := &models.Comment{
		TopicID:   1,
		Username:  "testuser",
		Content:   "Original content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, comment)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update comment
	newContent := "Updated content"
	comment.Content = newContent
	err = repo.Update(ctx, comment)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updated, err := repo.FindById(ctx, comment.Id)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}
	if updated.Content != newContent {
		t.Errorf("Expected content '%s', got '%s'", newContent, updated.Content)
	}
}

func TestCommentRepository_Delete(t *testing.T) {
	repo := NewCommentRepository(testDB)
	ctx := context.Background()

	// Create test comment
	comment := &models.Comment{
		TopicID:   1,
		Username:  "testuser",
		Content:   "To be deleted",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, comment)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Delete comment
	err = repo.Delete(ctx, comment.Id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	deleted, err := repo.FindById(ctx, comment.Id)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}
	if deleted != nil {
		t.Error("Expected comment to be deleted, but it was found")
	}
}

func TestCommentRepository_FindById_NotFound(t *testing.T) {
	repo := NewCommentRepository(testDB)
	ctx := context.Background()

	_, err := repo.FindById(ctx, 9999)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}
}

func TestCommentRepository_FindAll(t *testing.T) {
	repo := NewCommentRepository(testDB)
	ctx := context.Background()

	// Clear table
	_, err := testDB.Exec("TRUNCATE comments")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Create test data
	topicID := 1
	comments := []*models.Comment{
		{
			TopicID:   topicID,
			Username:  "user1",
			Content:   "Comment 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			TopicID:   topicID,
			Username:  "user2",
			Content:   "Comment 2",
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now().Add(-time.Hour),
		},
	}

	for _, c := range comments {
		err := repo.Create(ctx, c)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// Test FindAll
	found, err := repo.FindAll(ctx, topicID)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("Expected 2 comments, got %d", len(found))
	}

	// Verify order (should be DESC by created_at)
	if found[0].Content != "Comment 1" {
		t.Errorf("Expected first comment to be 'Comment 1', got '%s'", found[0].Content)
	}
}

func TestCommentRepository_FindAll_Empty(t *testing.T) {
	repo := NewCommentRepository(testDB)
	ctx := context.Background()

	// Clear table
	_, err := testDB.Exec("TRUNCATE comments")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test with no comments
	found, err := repo.FindAll(ctx, 1)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(found) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(found))
	}
}
