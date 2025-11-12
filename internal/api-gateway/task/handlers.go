package task

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oogway93/taskmanager/config"
	// auth "github.com/oogway93/taskmanager/internal/api-gateway/auth"

	AuthHandler "github.com/oogway93/taskmanager/internal/api-gateway/auth"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
)

type Handler struct {
	taskClient *Client
	authClient *AuthHandler.Client
	cfg        *config.Config
}

func NewHandler(cfg *config.Config, authClient *AuthHandler.Client) (*Handler, error) {
	client, err := NewClient(cfg.GetTaskGRPCAddress())
	if err != nil {
		return nil, err
	}

	return &Handler{
		taskClient: client,
		authClient: authClient,
		cfg:        cfg,
	}, nil
}

func (h *Handler) Create(c *gin.Context) {
	var req entity.TaskRequest

	// Валидация входных данных
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Invalid registration request", err)

		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid request data",
		})
		return
	}


	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}
	
	respAuth, err := h.authClient.GetUserProfile(userID.(string))
	if err != nil {
		// h.handleGRPCError(c, err)
		return
	}

	req.User_id = userID.(string)

	// Вызов gRPC сервиса аутентификации
	respTask, err := h.taskClient.CreateTask(req)
	if err != nil {
		// h.handleGRPCError(c, err)
		return
	}

	// Преобразование gRPC ответа в HTTP ответ
	response := entity.TaskResponse{
		Task: entity.Task{
			ID:          respTask.Task.Id,
			Title:       respTask.Task.Title,
			Description: respTask.Task.Description,
			Priority:    respTask.Task.Priority,
			Status:      respTask.Task.Status,
			Tags:        respTask.Task.Tags,
			User_id:     respAuth.User.Id,
			CreatedAt:   respTask.Task.CreatedAt.AsTime(),
			UpdatedAt:   respTask.Task.UpdatedAt.AsTime(),
		},
	}

	// logger.WithFields(logger.Fields{
	// 	"user_id": resp.User.Id,
	// 	"email":   resp.User.Email,
	// }).Info("User registered successfully")

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) Close() {
	h.taskClient.Close()
}
