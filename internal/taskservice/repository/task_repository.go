package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *entity.Task) error
}

type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository создает новый репозиторий пользователей
func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) CreateTask(ctx context.Context, task *entity.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, priority, status, user_id, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
	` //TODO: убрать raw sql, использовать gORM

	task.ID = uuid.New().String()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		task.ID,
		task.Title,
		task.Description,
		task.Priority,
		task.Status,
		task.User_id,
		pq.Array(task.Tags),
		task.CreatedAt,
		task.UpdatedAt,
	)

	return err
}
