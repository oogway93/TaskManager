package server

import (
	"context"
	"log"

	"github.com/oogway93/taskmanager/gen/task"
	"github.com/oogway93/taskmanager/internal/entity"
	"github.com/oogway93/taskmanager/internal/taskservice/service"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskServer struct {
	task.UnimplementedTaskServiceServer
	taskService service.TaskService
	Log         *zap.Logger
}

func NewTaskServer(taskService service.TaskService, Log *zap.Logger) *TaskServer {
	return &TaskServer{
		taskService: taskService,
		Log:         Log,
	}
}

func (s *TaskServer) CreateTask(ctx context.Context, req *task.Task) (*task.TaskResponse, error) {
	taskEntity := s.protoToTask(req)

	// Вызываем сервис
	createdTask, err := s.taskService.CreateTask(ctx, taskEntity)
	if err != nil {
		s.Log.Error("Error caused after calling the func Create Task", zap.Error(err))
		return nil, err
	}

	// Преобразуем результат обратно в protobuf
	return &task.TaskResponse{
		Task: s.taskToProto(createdTask),
	}, nil

}

func (s *TaskServer) ListTasks(ctx context.Context, req *task.ListTasksRequest) (*task.ListTasksResponse, error) {
	// Вызываем сервис
	tasks, err := s.taskService.ListTasks(ctx, req.UserId)
	if err != nil {
		s.Log.Error("Error caused after calling the func ListTasks", zap.Error(err))
		return nil, err
	}

	var listTasksProto []*task.Task
	for i := 0; i < len(tasks); i++ {
		listTasksProto = append(listTasksProto, s.taskToProto(&tasks[i]))
	}

	//Преобразуем результат обратно в protobuf
	return &task.ListTasksResponse{
		Tasks: listTasksProto,
		Total: int32(len(listTasksProto)),
	}, nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *task.GetTaskRequest) (*task.TaskResponse, error) {
	log.Println("input data from server grpc ", req.TaskId)
	taskSer, err := s.taskService.GetTask(ctx, req.TaskId)
	if err != nil {
		s.Log.Error("Error caused after calling the func GetTask", zap.Error(err))
		return nil, err
	}
	log.Println("task", taskSer)

	return &task.TaskResponse{
		Task: s.taskToProto(taskSer),
	}, nil
}

func (s *TaskServer) taskToProto(taskReq *entity.Task) *task.Task {
	return &task.Task{
		Id:          taskReq.ID,
		Title:       taskReq.Title,
		Description: taskReq.Description,
		Priority:    taskReq.Priority,
		Status:      taskReq.Status,
		UserId:      taskReq.User_id,
		Tags:        taskReq.Tags,
		CreatedAt:   timestamppb.New(taskReq.CreatedAt),
		UpdatedAt:   timestamppb.New(taskReq.UpdatedAt),
		// DueDate:     timestamppb.New(taskReq),
	}
}

func (s *TaskServer) protoToTask(taskProto *task.Task) *entity.Task {
	return &entity.Task{
		ID:          taskProto.Id,
		Title:       taskProto.Title,
		Description: taskProto.Description,
		Priority:    taskProto.Priority,
		Status:      taskProto.Status,
		User_id:     taskProto.UserId,
		Tags:        taskProto.Tags,
		CreatedAt:   taskProto.CreatedAt.AsTime(),
		UpdatedAt:   taskProto.UpdatedAt.AsTime(),
		// DueDate:     timestamppb.New(taskReq),
	}
}

// func (s *TaskServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
// 	// Создаем пользователя
// 	user, err := s.authService.Register(ctx, req.Email, req.Password, req.Name)
// 	if err != nil {
// 		return nil, status.Error(codes.Internal, err.Error())
// 	}

// 	// Генерируем токены
// 	accessToken, accessExp, err := s.tokenService.GenerateAccessToken(user)
// 	if err != nil {
// 		return nil, status.Error(codes.Internal, "failed to generate tokens")
// 	}

// 	refreshToken, _, err := s.tokenService.GenerateRefreshToken(user)
// 	if err != nil {
// 		return nil, status.Error(codes.Internal, "failed to generate tokens")
// 	}

// 	// logger.Info("User registered", "email", user.Email, "user_id", user.ID)
// 	log.Println("User registered", "email", user.Email, "user_id", user.ID)

// 	return &auth.RegisterResponse{
// 		AccessToken:  accessToken,
// 		RefreshToken: refreshToken,
// 		TokenType:    "Bearer",
// 		ExpiresAt:    timestamppb.New(accessExp),
// 		User:         s.userToProto(user),
// 	}, nil
// }
