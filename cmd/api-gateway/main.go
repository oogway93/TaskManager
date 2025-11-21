package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/internal/api-gateway"
	"github.com/oogway93/taskmanager/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	AuthHandler "github.com/oogway93/taskmanager/internal/api-gateway/auth"
	healthHandler "github.com/oogway93/taskmanager/internal/api-gateway/health"
	TaskHandler "github.com/oogway93/taskmanager/internal/api-gateway/task"
)

func main() {
	cfg := config.Load()

	Log := logger.Init(cfg)
	defer logger.Sync(Log)

	jwtConfig := middlewares.NewJWTConfig(cfg)

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	authHandler, err := AuthHandler.NewHandler(cfg, Log)
	if err != nil {
		Log.Fatal("Failed to create auth handler", zap.Error(err))
	}
	defer authHandler.Close()

	taskHandler, err := TaskHandler.NewHandler(cfg, authHandler.AuthClient, Log)
	if err != nil {
		Log.Fatal("Failed to create task handler", zap.Error(err))
	}
	defer taskHandler.Close()
	middlewares.PrometheusInit()
	router := gin.Default()
	router.Use(cors.Default())

	router.Use(middlewares.PrometheusMiddleware())

	router.Use(middlewares.JWTMiddleware(jwtConfig))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	public := router.Group("/api/v1")
	{
		public.GET("/health", healthHandler.HealthCheck)
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/registration", authHandler.Register)
	}

	protected := router.Group("/api/v1")
	{
		protected.GET("/auth/profile", authHandler.GetProfile)
		protected.POST("/task", taskHandler.Create)
		protected.GET("/task", taskHandler.ListTasks)
		protected.GET("/task/:id", taskHandler.GetTask)
	}

	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		Log.Info("🚀 API Gateway started on:", zap.String("address", cfg.GetServerAddress()))
		Log.Info("📊 Environment:", zap.String("env", cfg.App.Env))
		Log.Info("🔗 Task Service:", zap.String("Task Service URL", cfg.GetTaskServiceURL())) //TODO:подключить task и auth service в логи
		Log.Info("🔗 Auth Service:", zap.String("Auth Service URL", cfg.GetAuthServiceURL()))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Ожидание сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	Log.Info("🛑 Shutting down server...")

	// Даем время на завершение активных соединений
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	Log.Info("✅ Server exited properly")
}
