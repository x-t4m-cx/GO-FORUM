package postgres

import (
	"TopicService/internal/domain/models"
	"context"
	"database/sql"
	"errors"
	"time"
)

type TopicRepo interface {
	Create(ctx context.Context, topic *models.Topic) error
	Update(ctx context.Context, topic *models.Topic) error
	Delete(ctx context.Context, id int) error
	FindById(ctx context.Context, id int) (*models.Topic, error)
	FindAll(ctx context.Context) ([]*models.Topic, error)
}

type topicRepository struct {
	db *sql.DB
}

func NewTopicRepository(db *sql.DB) TopicRepo {
	return &topicRepository{db: db}
}

func (r *topicRepository) Create(ctx context.Context, topic *models.Topic) error {
	query := `INSERT INTO topics (title, content, username, created_at, updated_at) 
				VALUES ($1,$2,$3,$4,$5) RETURNING id;`
	return r.db.QueryRowContext(ctx, query,
		topic.Title, topic.Content,
		topic.Username, topic.CreatedAt, topic.UpdatedAt).Scan(&topic.Id)
}

func (r *topicRepository) Update(ctx context.Context, topic *models.Topic) error {
	query := `UPDATE topics 
				SET title = $1, content = $2, updated_at = $3
				WHERE id = $4;`
	_, err := r.db.ExecContext(ctx, query, topic.Title, topic.Content, time.Now(), topic.Id)
	return err
}

func (r *topicRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM topics WHERE id = $1;`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *topicRepository) FindById(ctx context.Context, id int) (*models.Topic, error) {
	query := `SELECT * FROM topics WHERE id = $1;`
	var topic models.Topic
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&topic.Id,
		&topic.Title,
		&topic.Content,
		&topic.Username,
		&topic.CreatedAt,
		&topic.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound{Entity: "Topic", Id: id}
		}
		return nil, err
	}
	return &topic, nil
}

func (r *topicRepository) FindAll(ctx context.Context) ([]*models.Topic, error) {
	query := `SELECT * FROM topics ORDER BY created_at DESC;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []*models.Topic
	for rows.Next() {
		var topic models.Topic
		if err := rows.Scan(
			&topic.Id,
			&topic.Title,
			&topic.Content,
			&topic.Username,
			&topic.CreatedAt,
			&topic.UpdatedAt); err != nil {
			return nil, err
		}
		topics = append(topics, &topic)
	}
	return topics, rows.Err()
}
