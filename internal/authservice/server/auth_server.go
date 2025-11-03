package server

import (
	"context"
	"log"

	"github.com/oogway93/taskmanager/gen/auth"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
	"github.com/oogway93/taskmanager/internal/authservice/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	authService  service.AuthService
	tokenService service.TokenService
}

func NewAuthServer(authService service.AuthService, tokenService service.TokenService) *AuthServer {
	return &AuthServer{
		authService:  authService,
		tokenService: tokenService,
	}
}

func (s *AuthServer) userToProto(user *entity.User) *auth.User {
    return &auth.User{
        Id:        user.ID,
        Email:     user.Email,
        Name:      user.Name,
        Role:      user.Role,
        Active:    user.Active,
        // Password НЕ включается! ✅
        CreatedAt: timestamppb.New(user.CreatedAt),
        UpdatedAt: timestamppb.New(user.UpdatedAt),
    }
}

func (s *AuthServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	// Создаем пользователя
	user, err := s.authService.Register(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Генерируем токены
	accessToken, accessExp, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	refreshToken, _, err := s.tokenService.GenerateRefreshToken(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	// logger.Info("User registered", "email", user.Email, "user_id", user.ID)
	log.Println("User registered", "email", user.Email, "user_id", user.ID)

	return &auth.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    timestamppb.New(accessExp),
		User:         s.userToProto(user),
	}, nil
}

