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
		Id:     user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		Active: user.Active,
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

func (s *AuthServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	log.Println("Login attempt", "email", req.Email)

	// Аутентифицируем пользователя
	user, err := s.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserNotFound, service.ErrInvalidCredentials:
			// logger.Warn("Login failed - invalid credentials", "email", req.Email)
			log.Println("Login failed - invalid credentials", "email", req.Email)
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case service.ErrUserInactive:
			log.Println("Login failed - inactive account", "email", req.Email)
			return nil, status.Error(codes.PermissionDenied, "account is deactivated")
		default:
			log.Println("Login failed - internal error", "error", err, "email", req.Email)
			return nil, status.Error(codes.Internal, "login failed")
		}
	}

	// Генерируем токены
	accessToken, accessExp, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		log.Println("Failed to generate access token", "error", err)
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	refreshToken, _, err := s.tokenService.GenerateRefreshToken(user)
	if err != nil {
		log.Println("Failed to generate refresh token", "error", err)
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	// // Сохраняем сессию (если используется Redis для сессий)
	// if err := s.authService.SaveSession(ctx, user.ID, refreshToken, refreshExp); err != nil {
	// 	logger.Warn("Failed to save session", "error", err)
	// 	// Не прерываем вход, если не удалось сохранить сессию
	// }

	log.Println("User logged in successfully", "email", user.Email, "user_id", user.ID)

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    timestamppb.New(accessExp),
		User:         s.userToProto(user),
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	claims, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		return &auth.ValidateTokenResponse{Valid: false}, nil
	}

	// Проверяем что пользователь все еще существует и активен
	user, err := s.authService.GetUserByID(ctx, claims.UserID)
	if err != nil || !user.Active {
		return &auth.ValidateTokenResponse{Valid: false}, nil
	}

	return &auth.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

func (s *AuthServer) GetUserProfile(ctx context.Context, req *auth.GetUserProfileRequest) (*auth.GetUserProfileResponse, error) {
	user, err := s.authService.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &auth.GetUserProfileResponse{
		User: s.userToProto(user),
	}, nil
}
