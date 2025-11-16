package service

import (
	"context"
	"errors"

	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
	"github.com/oogway93/taskmanager/internal/taskservice/repository"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserInactive       = errors.New("user is already inactive ")
)

type TaskService interface {
	CreateTask(ctx context.Context, task *entity.Task) (*entity.Task, error)
	ListTasks(ctx context.Context, userId string) ([]entity.Task, error)
}

type taskService struct {
	taskRepo repository.TaskRepository
	Log *zap.Logger
}

func NewTaskService(taskRepo repository.TaskRepository, Log *zap.Logger) TaskService {
	return &taskService{
		taskRepo: taskRepo,
		Log: Log, 
	}
}

func (s *taskService) CreateTask(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	if err := s.taskRepo.CreateTask(ctx, task); err != nil {
		s.Log.Fatal("Error caused, after calling repo's CreateTask, in task service",zap.Error(err))
		return nil, err
	}
	return task, nil
}

func (s *taskService) ListTasks(ctx context.Context, userId string) ([]entity.Task, error) {
	tasks, err := s.taskRepo.ListTasks(ctx, userId)
	return tasks, err 
}
