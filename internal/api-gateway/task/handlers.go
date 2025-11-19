package task

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oogway93/taskmanager/config"
	"go.uber.org/zap"

	AuthHandler "github.com/oogway93/taskmanager/internal/api-gateway/auth"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
)

type Handler struct {
	taskClient *Client
	authClient *AuthHandler.Client
	cfg        *config.Config
	Log        *zap.Logger
}

func NewHandler(cfg *config.Config, authClient *AuthHandler.Client, Log *zap.Logger) (*Handler, error) {
	client, err := NewClient(cfg.GetTaskGRPCAddress(), Log)
	if err != nil {
		return nil, err
	}

	return &Handler{
		taskClient: client,
		authClient: authClient,
		cfg:        cfg,
		Log:        Log,
	}, nil
}

func (h *Handler) Create(c *gin.Context) {
	var req entity.TaskRequest

	// Валидация входных данных
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.Error("Invalid Create task request", zap.Error(err))

		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid request data",
		})
		return
	}
	h.Log.Info("Данные о тегах из река", zap.Strings("tags", req.Tags))

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
		h.Log.Error("Error caused after calling func GetUserProfile in api-gateway task's handlers", zap.Error(err))
		return
	}

	req.User_id = userID.(string)

	// Вызов gRPC сервиса аутентификации
	respTask, err := h.taskClient.CreateTask(req)
	if err != nil {
		h.Log.Error("Error caused after calling func CreateTask in api-gateway task's handlers", zap.Error(err))
		return
	}

	// Преобразование gRPC ответа в HTTP ответ
	response := entity.TaskResponse{
		Task: &entity.Task{
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

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) ListTasks(c *gin.Context) {
	userID, existsUserID := c.Get("user_id")
	if !existsUserID {
		c.JSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}
	
	username, existsUsername := c.Get("username")
	if !existsUsername {
		h.Log.Error("Failed error in getting username from header")
		c.JSON(http.StatusUnauthorized, entity.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	respTask, err := h.taskClient.ListTasks(userID.(string))
	if err != nil {
		h.Log.Error("Error caused after calling func CreateTask in api-gateway task's handlers", zap.Error(err))
		return
	}
	var tasksEntity []*entity.TaskListData
	for i := 0; i < len(respTask.Tasks); i++ {
		taskEntity := &entity.TaskListData{
			Title:       respTask.Tasks[i].Title,
			Description: respTask.Tasks[i].Description,
			Priority:    respTask.Tasks[i].Priority,
			Status:      respTask.Tasks[i].Status,
			Tags:        respTask.Tasks[i].Tags,
			CreatedAt:   respTask.Tasks[i].CreatedAt.AsTime(),
			UpdatedAt:   respTask.Tasks[i].UpdatedAt.AsTime(),
		}
		tasksEntity = append(tasksEntity, taskEntity)
	}
	response := &entity.TaskListResponse{
		Tasks: tasksEntity,
		Total: respTask.Total,
		Username: username.(string),
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetTask(c *gin.Context) {
	taskId := c.Param("id")
	if taskId == "" {
		h.Log.Error("Nil Param 'id' in path from request")

		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid request path's id",
		})
	}

	req := entity.GetTaskRequest{
		TaskId: taskId,
	}

	respTask, err := h.taskClient.GetTask(req)
	if err != nil {
		h.Log.Fatal("Error caused after calling func CreateTask in api-gateway task's handlers", zap.Error(err))
		return
	}

	response := &entity.TaskResponse{
		Task: &entity.Task{
			ID:          respTask.Task.Id,
			Title:       respTask.Task.Title,
			Description: respTask.Task.Description,
			Priority:    respTask.Task.Priority,
			Status:      respTask.Task.Status,
			Tags:        respTask.Task.Tags,
			User_id:     respTask.Task.UserId,
			CreatedAt:   respTask.Task.CreatedAt.AsTime(),
			UpdatedAt:   respTask.Task.UpdatedAt.AsTime(),
		},
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) Close() {
	h.taskClient.Close()
}
