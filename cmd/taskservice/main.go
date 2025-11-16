package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/gen/task"
	"github.com/oogway93/taskmanager/internal/infrastructure/postgres"
	"github.com/oogway93/taskmanager/internal/taskservice/repository"
	"github.com/oogway93/taskmanager/internal/taskservice/server"
	"github.com/oogway93/taskmanager/internal/taskservice/service"
	"github.com/oogway93/taskmanager/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg := config.Load()

	Log := logger.Init(cfg)
	defer logger.Sync(Log)

	// Initialize database connection
	db, err := postgres.NewPostgresDB(cfg.GetDBConnectionString(), Log)
	if err != nil {
		Log.Fatal("Failed to connect to database:", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(db, Log)

	// Initialize services
	taskService := service.NewTaskService(taskRepo, Log)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	taskServer := server.NewTaskServer(taskService, Log)

	// Register auth service
	task.RegisterTaskServiceServer(grpcServer, taskServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", cfg.GetTaskGRPCAddress())
	if err != nil {
		Log.Fatal("Failed to listen an GRPC port:", zap.Error(err))
	}

	go func() {
		Log.Info("Task Service started on", zap.String("address", cfg.GetTaskGRPCAddress()))
		Log.Info("Environment:", zap.String("env", cfg.App.Env))

		if err := grpcServer.Serve(lis); err != nil {
			Log.Fatal("Failed to serve grpc server in Task Service:", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	Log.Info("Shutting down Task Service...")
	grpcServer.GracefulStop()
	Log.Info("Task Service stopped")
}
