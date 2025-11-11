package task

import (
	"fmt"

	"github.com/oogway93/taskmanager/gen/task"
	"github.com/oogway93/taskmanager/internal/api-gateway/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client task.TaskServiceClient
}

// NewClient создает новый клиент для работы с Task Service
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to task service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: task.NewTaskServiceClient(conn),
	}, nil
}

func (c *Client) CreateTask(taskReq entity.TaskRequest) (*task.TaskResponse, error) {
	return nil, nil
}
