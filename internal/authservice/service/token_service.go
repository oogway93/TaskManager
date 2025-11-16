package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// TokenClaims представляет данные, хранящиеся в JWT токене
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" или "refresh"
	jwt.RegisteredClaims
}

type TokenService interface {
	GenerateAccessToken(user *entity.User) (string, time.Time, error)
	GenerateRefreshToken(user *entity.User) (string, time.Time, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	GenerateTokenPair(user *entity.User) (accessToken, refreshToken string, accessExp, refreshExp time.Time, err error)
	ExtractUserIDFromToken(tokenString string) (string, error)
	RefreshTokenPair(refreshToken string, user *entity.User) (newAccessToken, newRefreshToken string, accessExp, refreshExp time.Time, err error)
	IsTokenExpired(tokenString string) bool
}

// TokenService отвечает за создание и валидацию JWT токенов
type tokenService struct {
	cfg *config.Config
	Log *zap.Logger
}

// NewTokenService создает новый сервис для работы с токенами
func NewTokenService(cfg *config.Config, Log *zap.Logger) TokenService {
	return &tokenService{
		cfg: cfg,
		Log: Log,
	}
}

// GenerateAccessToken создает access token для пользователя
func (s *tokenService) GenerateAccessToken(user *entity.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.cfg.JWT.AccessTTL)

	claims := TokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "taskmanager-auth",
			Subject:   user.ID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		s.Log.Fatal("Error caused after calling func SignedString in tokenservice", zap.Error(err))
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

// GenerateRefreshToken создает refresh token для пользователя
func (s *tokenService) GenerateRefreshToken(user *entity.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.cfg.JWT.RefreshTTL)

	claims := TokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "taskmanager-auth",
			Subject:   user.ID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		s.Log.Fatal("Error caused after calling func SignedString in tokenservice", zap.Error(err))
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

// ValidateToken проверяет валидность токена и возвращает claims
func (s *tokenService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			s.Log.Fatal("Error caused after calling func ParseWithClaims in tokenservice")
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		var validationErr *jwt.ValidationError
		if errors.As(err, &validationErr) {
			if validationErr.Errors&jwt.ValidationErrorExpired != 0 {
				s.Log.Fatal("Error caused after calling func ParseWithClaims in tokenservice", zap.Error(err))
				return nil, ErrExpiredToken
			}
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		s.Log.Fatal("Error caused after calling func Claims in tokenservice", zap.Error(err))
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// GenerateTokenPair создает пару access и refresh токенов
func (s *tokenService) GenerateTokenPair(user *entity.User) (accessToken, refreshToken string, accessExp, refreshExp time.Time, err error) {
	accessToken, accessExp, err = s.GenerateAccessToken(user)
	if err != nil {
		s.Log.Fatal("Error caused after calling func GenerateTokenPair in tokenservice", zap.Error(err))
		return "", "", time.Time{}, time.Time{}, err
	}

	refreshToken, refreshExp, err = s.GenerateRefreshToken(user)
	if err != nil {
		s.Log.Fatal("Error caused after calling func GenerateRefreshPair in tokenservice", zap.Error(err))
		return "", "", time.Time{}, time.Time{}, err
	}

	return accessToken, refreshToken, accessExp, refreshExp, nil
}

// ExtractUserIDFromToken извлекает ID пользователя из токена (без полной валидации)
func (s *tokenService) ExtractUserIDFromToken(tokenString string) (string, error) {
	// Парсим токен без проверки подписи (только для извлечения данных)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &TokenClaims{})
	if err != nil {
		s.Log.Fatal("Error caused after calling func ParseUnverified in tokenservice", zap.Error(err))
		return "", ErrInvalidToken
	}

	if claims, ok := token.Claims.(*TokenClaims); ok {
		s.Log.Fatal("Error caused after calling func Claims in tokenservice", zap.Error(err))
		return claims.UserID, nil
	}

	return "", ErrInvalidToken
}

// IsTokenExpired проверяет истек ли срок действия токена
func (s *tokenService) IsTokenExpired(tokenString string) bool {
	_, err := s.ValidateToken(tokenString)
	return errors.Is(err, ErrExpiredToken)
}

// RefreshTokenPair обновляет пару токенов
func (s *tokenService) RefreshTokenPair(refreshToken string, user *entity.User) (newAccessToken, newRefreshToken string, accessExp, refreshExp time.Time, err error) {
	// Валидируем refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		s.Log.Fatal("Error caused after calling func ValidateToken in tokenservice", zap.Error(err))
		return "", "", time.Time{}, time.Time{}, err
	}

	// Проверяем что это действительно refresh token
	if claims.TokenType != "refresh" {
		s.Log.Fatal("Error caused in make check tokenType in tokenservice", zap.Error(err))
		return "", "", time.Time{}, time.Time{}, ErrInvalidToken
	}

	// Проверяем что токен принадлежит тому же пользователю
	if claims.UserID != user.ID {
		s.Log.Fatal("Error caused in make check UserID in tokenservice", zap.Error(err))
		return "", "", time.Time{}, time.Time{}, ErrInvalidToken
	}

	// Генерируем новую пару токенов
	return s.GenerateTokenPair(user)
}
