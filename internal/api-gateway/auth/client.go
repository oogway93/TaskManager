package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/oogway93/taskmanager/gen/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client auth.AuthServiceClient
}

// NewClient создает новый клиент для работы с Auth Service
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: auth.NewAuthServiceClient(conn),
	}, nil
}

// Close закрывает соединение с сервером
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Register регистрирует нового пользователя
func (c *Client) Register(email, password, username string) (*auth.RegisterResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req := &auth.RegisterRequest{
		Email:    email,
		Password: password,
		Username: username,
	}

	resp, err := c.client.Register(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) Login(email, password string) (*auth.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req := &auth.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.client.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) ValidateToken(token string) (*auth.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &auth.ValidateTokenRequest{
		Token: token,
	}

	resp, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		return nil, err
		// return nil, wrapGRPCError(err)
	}

	return resp, nil
}

func (c *Client) GetUserProfile(userID string) (*auth.GetUserProfileResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &auth.GetUserProfileRequest{
		UserId: userID,
	}

	resp, err := c.client.GetUserProfile(ctx, req)
	if err != nil {
		// return nil, wrapGRPCError(err)
		return nil, err
	}

	return resp, nil
}
