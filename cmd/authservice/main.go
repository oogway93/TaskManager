package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/gen/auth"
	"github.com/oogway93/taskmanager/internal/authservice/repository"
	"github.com/oogway93/taskmanager/internal/authservice/server"
	"github.com/oogway93/taskmanager/internal/authservice/service"
	"github.com/oogway93/taskmanager/internal/infrastructure/postgres"
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
	userRepo := repository.NewUserRepository(db, Log)

	// Initialize services
	tokenService := service.NewTokenService(cfg, Log)
	authService := service.NewAuthService(userRepo, tokenService, Log)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	authServer := server.NewAuthServer(authService, tokenService, Log)

	// Register auth service
	auth.RegisterAuthServiceServer(grpcServer, authServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", cfg.GetAuthGRPCAddress())
	if err != nil {
		Log.Fatal("Failed to listen grpc's port:", zap.Error(err))
	}

	go func() {
		Log.Info("Auth Service started on ", zap.String("address", cfg.GetAuthGRPCAddress()))
		Log.Info("Environment:", zap.String("env", cfg.App.Env))
		Log.Info("JWT Access TTL:", zap.Duration("ttl", cfg.JWT.AccessTTL))

		if err := grpcServer.Serve(lis); err != nil {
			Log.Fatal("Failed to serve an gRPC server:", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	Log.Info("Shutting down Auth Service...")
	grpcServer.GracefulStop()
	Log.Info("Auth Service stopped")
}
