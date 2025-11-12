package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
)

type Handler struct {
	AuthClient *Client
	cfg        *config.Config
}

func NewHandler(cfg *config.Config) (*Handler, error) {
	client, err := NewClient(cfg.GetAuthGRPCAddress())
	if err != nil {
		return nil, err
	}

	return &Handler{
		AuthClient: client,
		cfg:        cfg,
	}, nil
}

func (h *Handler) Register(c *gin.Context) {
	var req entity.RegisterRequest

	// Валидация входных данных
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Invalid registration request", err)

		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid request data",
		})
		return
	}

	// // Дополнительная валидация
	// if err := h.validateRegisterRequest(&req); err != nil {
	// 	c.JSON(http.StatusBadRequest, ErrorResponse{
	// 		Error:   "VALIDATION_ERROR",
	// 		Message: err.Error(),
	// 	})
	// 	return
	// }

	// Вызов gRPC сервиса аутентификации
	resp, err := h.AuthClient.Register(req.Email, req.Password, req.Name)
	if err != nil {
		// h.handleGRPCError(c, err)
		return
	}

	// Преобразование gRPC ответа в HTTP ответ
	response := entity.RegisterResponse{
		AccessToken:  resp.AccessToken, //TODO:убрать вывод обратно пользователю и генерацию токена при регистрации
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
		ExpiresAt:    resp.ExpiresAt.AsTime(),
		User: entity.UserResponse{
			ID:        resp.User.Id,
			Email:     resp.User.Email,
			Name:      resp.User.Name,
			Role:      resp.User.Role,
			CreatedAt: resp.User.CreatedAt.AsTime(),
		},
	}

	// logger.WithFields(logger.Fields{
	// 	"user_id": resp.User.Id,
	// 	"email":   resp.User.Email,
	// }).Info("User registered successfully")

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) Login(c *gin.Context) {
	var req entity.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Invalid login request:::", err)

		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid request data",
		})
		return
	}

	resp, err := h.AuthClient.Login(req.Email, req.Password)
	if err != nil {
		// h.handleGRPCError(c, err)
		return
	}

	response := entity.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
		ExpiresAt:    resp.ExpiresAt.AsTime(),
		User: entity.UserResponse{
			ID:        resp.User.Id,
			Email:     resp.User.Email,
			Name:      resp.User.Name,
			Role:      resp.User.Role,
			CreatedAt: resp.User.CreatedAt.AsTime(),
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	resp, err := h.AuthClient.GetUserProfile(userID.(string))
	if err != nil {
		// h.handleGRPCError(c, err)
		return
	}

	response := entity.UserResponse{
		ID:        resp.User.Id,
		Email:     resp.User.Email,
		Name:      resp.User.Name,
		Role:      resp.User.Role,
		CreatedAt: resp.User.CreatedAt.AsTime(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Close() {
	h.AuthClient.Close()
}
