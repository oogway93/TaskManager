package task

import "github.com/oogway93/taskmanager/config"

type Handler struct {
	taskClient *Client
	cfg        *config.Config
}

func NewHandler(cfg *config.Config) (*Handler, error) {
	client, err := NewClient(cfg.GetGRPCAddress())
	if err != nil {
		return nil, err
	}

	return &Handler{
		taskClient: client,
		cfg:        cfg,
	}, nil
}

