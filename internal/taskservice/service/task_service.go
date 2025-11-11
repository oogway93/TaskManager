package service

import (
	"errors"

	"github.com/oogway93/taskmanager/internal/taskservice/repository"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserInactive       = errors.New("user is already inactive ")
)

type TaskService interface {
	// Register(ctx context.Context, email, password, name string) (*entity.User, error)
	// Login(ctx context.Context, email, password string) (*entity.User, error)
	// ValidateToken(token string) (*TokenClaims, error)
	// GetUserByID(ctx context.Context, userID string) (*entity.User, error)
	// UpdateUserProfile(ctx context.Context, userID, email, name string) (*entity.User, error)
}

type taskService struct {
	taskRepo     repository.TaskRepository
}

func NewTaskService(userRepo repository.TaskRepository) TaskService {
	return &taskService{
		taskRepo:     userRepo,
	}
}

// func (s *authService) Register(ctx context.Context, email, password, name string) (*entity.User, error) {
// 	// Проверяем существует ли пользователь
// 	existing, _ := s.userRepo.GetByEmail(ctx, email)
// 	if existing != nil {
// 		return nil, ErrUserAlreadyExists
// 	}

// 	// Создаем пользователя
// 	user := &entity.User{
// 		Email:     email,
// 		Password:  password,
// 		Name:      name,
// 		Role:      "user",
// 		Active:    true,
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}

// 	// Хэшируем пароль
// 	hashedPassword, err := hashPassword(user.Password)
// 	if err != nil {
// 		return nil, err
// 	}
// 	user.Password = string(hashedPassword)

// 	// Сохраняем в БД
// 	if err := s.userRepo.Create(ctx, user); err != nil {
// 		return nil, err
// 	}

// 	return user, nil
// }

// func (s *authService) Login(ctx context.Context, email, password string) (*entity.User, error) {
// 	// Получаем пользователя по email
// 	user, err := s.userRepo.GetByEmail(ctx, email)
// 	if err != nil {
// 		return nil, ErrUserNotFound
// 	}

// 	// Проверяем пароль
// 	if !checkPassword(user.Password, password) {
// 		return nil, ErrInvalidCredentials
// 	}

// 	// Проверяем активность аккаунта
// 	if !user.Active {
// 		return nil, errors.New("account is deactivated")
// 	}

// 	// Обновляем время последнего входа (опционально)
// 	user.UpdatedAt = time.Now()
// 	// if err := s.userRepo.Update(ctx, user); err != nil {
// 	// 	// Логируем ошибку, но не прерываем вход
// 	// 	// logger.WithError(err).Warn("Failed to update last login time")
// 	// 	log.Println("Failed to update last login time")
// 	// }

// 	return user, nil
// }

// func (s *authService) ValidateToken(token string) (*TokenClaims, error) {
// 	return s.tokenService.ValidateToken(token)
// }

// func (s *authService) GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
// 	return s.userRepo.GetByID(ctx, userID)
// }

// func hashPassword(userPassword string) ([]byte, error) {
// 	hashed, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return hashed, nil
// }

// func checkPassword(hashedPassword, password string) bool {
// 	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
// 	return err == nil
// }
