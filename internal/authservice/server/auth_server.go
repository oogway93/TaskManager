package server

import (
	"context"

	"github.com/oogway93/taskmanager/gen/auth"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
	"github.com/oogway93/taskmanager/internal/authservice/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	authService  service.AuthService
	tokenService service.TokenService
	Log          *zap.Logger
}

func NewAuthServer(authService service.AuthService, tokenService service.TokenService, Log *zap.Logger) *AuthServer {
	return &AuthServer{
		authService:  authService,
		tokenService: tokenService,
		Log:          Log,
	}
}

func (s *AuthServer) userToProto(user *entity.User) *auth.User {
	return &auth.User{
		Id:     user.ID,
		Email:  user.Email,
		Username:   user.Username,
		Role:   user.Role,
		Active: user.Active,
		// Password НЕ включается! ✅
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func (s *AuthServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	// Создаем пользователя
	user, err := s.authService.Register(ctx, req.Email, req.Password, req.Username)
	if err != nil {
		s.Log.Error("Error caused after calling func Register from auth service", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Генерируем токены
	accessToken, accessExp, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		s.Log.Error("Error caused after calling func GenerateAccessToken from auth service", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	refreshToken, _, err := s.tokenService.GenerateRefreshToken(user)
	if err != nil {
		s.Log.Error("Error caused after calling func GenerateRefreshToken from auth service", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	s.Log.Info("User registered", zap.String("email", user.Email), zap.String("user_id", user.ID))

	return &auth.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    timestamppb.New(accessExp),
		User:         s.userToProto(user),
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	s.Log.Info("Login attempt", zap.String("email", req.Email))

	// Аутентифицируем пользователя
	user, err := s.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserNotFound, service.ErrInvalidCredentials:
			s.Log.Fatal("Login failed - invalid credentials", zap.String("email", req.Email))
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case service.ErrUserInactive:
			s.Log.Fatal("Login failed - inactive account", zap.String("email", req.Email))
			return nil, status.Error(codes.PermissionDenied, "account is deactivated")
		default:
			s.Log.Fatal("Login failed - internal error", zap.Error(err), zap.String("email", req.Email))
			return nil, status.Error(codes.Internal, "login failed")
		}
	}

	// Генерируем токены
	accessToken, accessExp, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		s.Log.Fatal("Failed to generate access token", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	refreshToken, _, err := s.tokenService.GenerateRefreshToken(user)
	if err != nil {
		s.Log.Fatal("Failed to generate refresh token", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	// // Сохраняем сессию (если используется Redis для сессий)
	// if err := s.authService.SaveSession(ctx, user.ID, refreshToken, refreshExp); err != nil {
	// 	logger.Warn("Failed to save session", "error", err)
	// 	// Не прерываем вход, если не удалось сохранить сессию
	// }

	s.Log.Info("User logged in successfully", zap.String("email", user.Email), zap.String("user_id", user.ID))

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
		s.Log.Fatal("Error caused after trying func ValidateToken", zap.Error(err))
		return &auth.ValidateTokenResponse{Valid: false}, nil
	}

	// Проверяем что пользователь все еще существует и активен
	user, err := s.authService.GetUserByID(ctx, claims.UserID)
	if err != nil || !user.Active {
		s.Log.Fatal("Error caused after calling func GetUserByID or user is not active", zap.Error(err))
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
		s.Log.Error("Error caused after calling GetUserByID", zap.Error(err))
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &auth.GetUserProfileResponse{
		User: s.userToProto(user),
	}, nil
}
