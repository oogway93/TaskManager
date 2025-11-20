package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/lib/pq"
	"github.com/oogway93/taskmanager/internal/entity"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *entity.Task) error
	ListTasks(ctx context.Context, userId string) ([]entity.Task, error)
	GetTaskByID(ctx context.Context, taskId uuid.UUID) (entity.Task, error)
}

type taskRepository struct {
	db  *sql.DB
	Log *zap.Logger
}

// NewTaskRepository создает новый репозиторий пользователей
func NewTaskRepository(db *sql.DB, Log *zap.Logger) TaskRepository {
	return &taskRepository{db: db, Log: Log}
}

func (r *taskRepository) CreateTask(ctx context.Context, task *entity.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, priority, status, user_id, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
	` //TODO: убрать raw sql, использовать gORM
	randomUUID, err := uuid.NewV4()
	if err != nil {
		r.Log.Error("Failed generate random UUID", zap.Error(err))
		return err
	}
	task.ID = randomUUID.String()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
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

func (r *taskRepository) ListTasks(ctx context.Context, userId string) ([]entity.Task, error) {
	query := `
	SELECT id, title, description, priority, status, tags, user_id, created_at, updated_at 
    FROM tasks WHERE user_id = $1;	
	`
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []entity.Task
	for rows.Next() {
		var task entity.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Priority,
			&task.Status,
			pq.Array(&task.Tags), // Используем pq.Array для сканирования массива
			&task.User_id,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		r.Log.Fatal("SQL error caused in repo's ListTasks", zap.Error(err))
		return nil, ErrUserNotFound
	}

	return tasks, nil
}

func (r *taskRepository) GetTaskByID(ctx context.Context, taskId uuid.UUID) (entity.Task, error) {
	query := `
	SELECT id, title, description, priority, status, tags, user_id, created_at, updated_at 
    FROM tasks WHERE id = $1;	
	`
	row, err := r.db.QueryContext(ctx, query, taskId)
	if err != nil {
		return entity.Task{}, err
	}
	defer row.Close()

	var task entity.Task
	for row.Next() {
		err = row.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Priority,
			&task.Status,
			pq.Array(&task.Tags),
			&task.User_id,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
	}

	if err = row.Err(); err != nil {
		r.Log.Error("SQL error caused in repo's GetTask", zap.Error(err))
		return entity.Task{}, ErrUserNotFound
	}
	r.Log.Info("data from repo", zap.String("id", task.ID), zap.String("title", task.Title))
	return task, nil
}
