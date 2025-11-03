package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/gen/auth"
	"github.com/oogway93/taskmanager/internal/authservice/repository"
	"github.com/oogway93/taskmanager/internal/authservice/server"
	"github.com/oogway93/taskmanager/internal/authservice/service"
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
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	tokenService := service.NewTokenService(cfg)
	authService := service.NewAuthService(userRepo, tokenService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	authServer := server.NewAuthServer(authService, tokenService)
	
	// Register auth service
	auth.RegisterAuthServiceServer(grpcServer, authServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", cfg.GetGRPCAddress())
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Printf("Auth Service started on %s", cfg.GetGRPCAddress())
		log.Printf("Environment: %s", cfg.App.Env)
		log.Printf("JWT Access TTL: %v", cfg.JWT.AccessTTL)
		
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Auth Service...")
	grpcServer.GracefulStop()
	log.Println("Auth Service stopped")
}