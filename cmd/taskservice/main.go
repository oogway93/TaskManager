package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/gen/task"
	"github.com/oogway93/taskmanager/internal/taskservice/repository"
	"github.com/oogway93/taskmanager/internal/taskservice/server"
	"github.com/oogway93/taskmanager/internal/taskservice/service"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := repository.NewPostgresDB(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(db)

	// Initialize services
	taskService := service.NewTaskService(taskRepo)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	taskServer := server.NewTaskServer(taskService)

	// Register auth service
	task.RegisterTaskServiceServer(grpcServer, taskServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", cfg.GetTaskGRPCAddress())
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Printf("Task Service started on %s", cfg.GetTaskGRPCAddress())
		log.Printf("Environment: %s", cfg.App.Env)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Task Service...")
	grpcServer.GracefulStop()
	log.Println("Task Service stopped")
}
