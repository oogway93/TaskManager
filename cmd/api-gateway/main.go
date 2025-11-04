package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/logger"
	"go.uber.org/zap"

	AuthHandler "github.com/oogway93/taskmanager/internal/api-gateway/auth"
	healthHandler "github.com/oogway93/taskmanager/internal/api-gateway/health"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg)
	defer logger.Sync()

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	authHandler, err := AuthHandler.NewHandler(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to create auth handler", zap.Error(err))
	}
	defer authHandler.Close()
	router := gin.Default()
	public := router.Group("/api/v1")
	{
		public.GET("/health", healthHandler.HealthCheck)
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/registration", authHandler.Register)
		public.GET("/auth/logout", healthHandler.HealthCheck)
	}
	// protected := router.Group("/api/v1") //TODO: написать protected routes, защищенные middleware через JWT token
	// router.Use(middleware.Auth())

	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Log.Info("🚀 API Gateway started on:", zap.String("address", cfg.GetServerAddress()))
		logger.Log.Info("📊 Environment:", zap.String("env", cfg.App.Env))
		// log.Printf("🔗 Task Service: %s", cfg.TaskServiceURL) //TODO:подключить task и auth service в логи
		// log.Printf("🔗 Auth Service: %s", cfg.AuthServiceURL)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Failed to start server", zap.Error(err))
		}
	}()

	// Ожидание сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("🛑 Shutting down server...")

	// Даем время на завершение активных соединений
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Debug("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("✅ Server exited properly")
}
