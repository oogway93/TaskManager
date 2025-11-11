package server

import (
	"github.com/oogway93/taskmanager/gen/task"
	"github.com/oogway93/taskmanager/internal/taskservice/service"
)

type TaskServer struct {
	task.UnimplementedTaskServiceServer	
	taskService  service.TaskService
}

func NewTaskServer(taskService service.TaskService) *TaskServer {
	return &TaskServer{
		taskService: taskService,
	}
}

// func (s *TaskServer) userToProto(user *entity.User) *auth.User {
// 	return &auth.User{
// 		Id:     user.ID,
// 		Email:  user.Email,
// 		Name:   user.Name,
// 		Role:   user.Role,
// 		Active: user.Active,
// 		// Password НЕ включается! ✅
// 		CreatedAt: timestamppb.New(user.CreatedAt),
// 		UpdatedAt: timestamppb.New(user.UpdatedAt),
// 	}
// }

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
