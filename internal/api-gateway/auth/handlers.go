package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
	"go.uber.org/zap"
)

type Handler struct {
	AuthClient *Client
	cfg        *config.Config
	Log        *zap.Logger
}

func NewHandler(cfg *config.Config, Log *zap.Logger) (*Handler, error) {
	client, err := NewClient(cfg.GetAuthGRPCAddress())
	if err != nil {
		return nil, err
	}

	return &Handler{
		AuthClient: client,
		cfg:        cfg,
		Log:        Log,
	}, nil
}

func (h *Handler) Register(c *gin.Context) {
	var req entity.RegisterRequest

	// Валидация входных данных
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.Error("Invalid registration request", zap.Error(err))

		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid request data",
		})
		return
	}

	// Вызов gRPC сервиса аутентификации
	resp, err := h.AuthClient.Register(req.Email, req.Password, req.Username)
	if err != nil {
		h.Log.Error("Error caused after calling func Register in api-gateway auth's handlers", zap.Error(err))
		return
	}

	// Преобразование gRPC ответа в HTTP ответ
	response := entity.RegisterResponse{
		Status: "User Registered",
		User: entity.UserResponse{
			Email:     resp.User.Email,
			Username:  resp.User.Username,
			Role:      resp.User.Role,
			CreatedAt: resp.User.CreatedAt.AsTime(),
		},
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) Login(c *gin.Context) {
	var req entity.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.Fatal("Invalid login request", zap.Error(err))

		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid request data",
		})
		return
	}

	resp, err := h.AuthClient.Login(req.Email, req.Password)
	if err != nil {
		h.Log.Fatal("Error caused after calling auth's client func Login in api-gateway auth's handlers", zap.Error(err))
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
			Username:  resp.User.Username,
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
		h.Log.Fatal("Error caused after calling auth's client func GetUserProfile in api-gateway auth's handlers", zap.Error(err))
		return
	}

	response := entity.UserResponse{
		ID:        resp.User.Id,
		Email:     resp.User.Email,
		Username:  resp.User.Username,
		Role:      resp.User.Role,
		CreatedAt: resp.User.CreatedAt.AsTime(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Close() {
	h.AuthClient.Close()
}
