package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/oogway93/taskmanager/internal/authservice/repository"
	"github.com/oogway93/taskmanager/internal/entity"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserInactive       = errors.New("user is already inactive ")
)

type AuthService interface {
	Register(ctx context.Context, email, password, name string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
	ValidateToken(token string) (*TokenClaims, error)
	GetUserByID(ctx context.Context, userID string) (*entity.User, error)
}

type authService struct {
	userRepo     repository.UserRepository
	tokenService TokenService
	Log          *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, tokenService TokenService, Log *zap.Logger) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenService: tokenService,
		Log:          Log,
	}
}

func (s *authService) Register(ctx context.Context, email, password, username string) (*entity.User, error) {
	// Проверяем существует ли пользователь
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		s.Log.Error("Error caused after trying repo's GetByEmail", zap.Error(err))
		return nil, ErrUserAlreadyExists
	}

	// Создаем пользователя
	user := &entity.User{
		Email:     email,
		Password:  password,
		Username:  username,
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Хэшируем пароль
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		s.Log.Error("Error caused after calling the func HashPassword", zap.Error(err))
		return nil, err
	}
	user.Password = string(hashedPassword)

	// Сохраняем в БД
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.Log.Error("Error caused after trying repo's Create in Auth Service", zap.Error(err))
		return nil, err
	}
	sendVerificationEmail(user.Email)

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*entity.User, error) {
	// Получаем пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.Log.Fatal("Error caused after trying repo's GetByEmail in Auth Service", zap.Error(err))
		return nil, ErrUserNotFound
	}

	// Проверяем пароль
	if !checkPassword(user.Password, password) {
		s.Log.Fatal("Error caused after trying CheckPassword in Auth Service")
		return nil, ErrInvalidCredentials
	}

	// Проверяем активность аккаунта
	if !user.Active {
		s.Log.Fatal("Error caused after making check of user isActive in Auth Service")
		return nil, errors.New("account is deactivated")
	}

	// Обновляем время последнего входа (опционально)
	// user.UpdatedAt = time.Now()
	// if err := s.userRepo.Update(ctx, user); err != nil {
	// 	// Логируем ошибку, но не прерываем вход
	// 	// logger.WithError(err).Warn("Failed to update last login time")
	// 	log.Println("Failed to update last login time")
	// }

	return user, nil
}

func (s *authService) ValidateToken(token string) (*TokenClaims, error) {
	return s.tokenService.ValidateToken(token)
}

func (s *authService) GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
	uuidUserID, err := uuid.Parse(userID)
	if err != nil {
		s.Log.Error("Failed error in trying parse userID string's type to UUID", zap.Error(err))
	}
	return s.userRepo.GetByID(ctx, uuidUserID)
}

func hashPassword(userPassword string) ([]byte, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hashed, nil
}

func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
func sendVerificationEmail(email string) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Ошибка подключения к RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Ошибка открытия канала: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"email_greetings", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		log.Fatalf("Ошибка объявления очереди: %s", err)
	}

	message := entity.EmailMessage{
		EmailTo: email,
	}

	body, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Ошибка сериализации сообщения: %s", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		log.Fatalf("Ошибка отправки сообщения: %s", err)
	}

	log.Printf("Отправлен email с кодом для: %s", email)
}
