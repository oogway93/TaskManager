package entity

import (
	"time"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Username string `json:"username" binding:"required,min=2,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

type UserInfo struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID        string
	Email     string
	Password  string
	Username  string
	Role      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RegisterResponse struct {
	// AccessToken  string       `json:"access_token"`
	// RefreshToken string       `json:"refresh_token"`
	// TokenType    string       `json:"token_type"`
	// ExpiresAt    time.Time    `json:"expires_at"`
	Status string       `json:"status"`
	User   UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string    `json:"id,omitempty"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Task struct {
	ID          string
	Title       string
	Description string
	Priority    string
	Status      string
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DueDate     time.Time
	User_id     string
}

// type TaskCreate struct {
// 	Title       string   `json:"title"`
// 	Description string   `json:"description"`
// 	Priority    string   `json:"priority"` //TODO:сделать enum, чтобы проверялось правильность введения
// 	Status      string   `json:"status"`
// 	Tags        []string `json:"tags"`
// 	User_id     string
// }

type TaskRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"` //TODO:сделать enum, чтобы проверялось правильность введения
	Status      string   `json:"status"`
	Tags        []string `json:"tags"`
	User_id     string
}

type TaskResponse struct {
	Task *Task `json:"task"`
}

type TaskListData struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"` //TODO:сделать enum, чтобы проверялось правильность введения
	Status      string   `json:"status"`
	Tags        []string `json:"tags"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskListResponse struct {
	Tasks    []*TaskListData `json:"tasks"`
	Total    int32           `json:"total"`
	Username string          `json:"username"`
}

type GetTaskRequest struct {
	TaskId string
}
