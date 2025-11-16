package task

import (
	"context"
	"fmt"
	"time"

	"github.com/oogway93/taskmanager/gen/task"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client task.TaskServiceClient
	Log    *zap.Logger
}

// NewClient создает новый клиент для работы с Task Service
func NewClient(serverAddr string, Log *zap.Logger) (*Client, error) {
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to task service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: task.NewTaskServiceClient(conn),
		Log:    Log,
	}, nil
}

func (c *Client) CreateTask(taskReq entity.TaskRequest) (*task.TaskResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req := &task.Task{
		Title:       taskReq.Title,
		Description: taskReq.Description,
		Priority:    taskReq.Priority,
		UserId:      taskReq.User_id,
	}

	resp, err := c.client.CreateTask(ctx, req)
	if err != nil {
		c.Log.Fatal("Error caused in Create task client", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListTasks(userId string) (*task.ListTasksResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req := &task.ListTasksRequest{UserId: userId}

	resp, err := c.client.ListTasks(ctx, req)
	if err != nil {
		c.Log.Fatal("Error caused in ListTasks task's client", zap.Error(err))
		return nil, err
	}

	
	

	return resp, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
