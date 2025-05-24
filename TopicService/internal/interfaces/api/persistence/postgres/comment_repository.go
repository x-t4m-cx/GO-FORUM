package postgres

import (
	"TopicService/internal/domain/models"
	"context"
	"database/sql"
	"errors"
	"time"
)

type CommentRepo interface {
	Create(ctx context.Context, comment *models.Comment) error
	Update(ctx context.Context, comment *models.Comment) error
	Delete(ctx context.Context, id int) error
	FindById(ctx context.Context, id int) (*models.Comment, error)
	FindAll(ctx context.Context, topicId int) ([]*models.Comment, error)
}

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepo {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	query := `INSERT INTO comments (topic_id, username, content, created_at, updated_at) 
				VALUES ($1,$2,$3,$4,$5) RETURNING id;`
	return r.db.QueryRowContext(ctx, query,
		comment.TopicID, comment.Username,
		comment.Content, comment.CreatedAt, comment.UpdatedAt).Scan(&comment.Id)
}

func (r *commentRepository) Update(ctx context.Context, comment *models.Comment) error {
	query := `UPDATE comments 
				SET content = $1, updated_at = $2
				WHERE id = $3;`
	_, err := r.db.ExecContext(ctx, query, comment.Content, time.Now(), comment.Id)
	return err
}

func (r *commentRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM comments WHERE id = $1;`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *commentRepository) FindById(ctx context.Context, id int) (*models.Comment, error) {
	query := `SELECT * FROM comments WHERE id = $1;`
	var comment models.Comment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.Id, &comment.TopicID,
		&comment.Username, &comment.Content,
		&comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound{Entity: "Comment", Id: id}
		}
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) FindAll(ctx context.Context, topicId int) ([]*models.Comment, error) {
	query := `SELECT * FROM comments WHERE topic_id = $1 ORDER BY created_at DESC;`
	rows, err := r.db.QueryContext(ctx, query, topicId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.Id, &comment.TopicID,
			&comment.Username, &comment.Content,
			&comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}
	return comments, rows.Err()
}
