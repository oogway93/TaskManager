package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	App    App
	JWT    JWTConfig
	DB     DBConfig
	Email  EmailConfig
}

type App struct {
	Env string
}

type ServerConfig struct {
	Host string
	Port int
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type EmailConfig struct {
	EmailFrom string
	EmailPass string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return &Config{
		ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnvInt("SERVER_PORT", 8000),
		},
		App{
			Env: getEnv("APP_ENV", "development"),
		},
		JWTConfig{
			Secret:     getEnv("JWT_SECRET", "asd"),
			AccessTTL:  time.Duration(getEnvInt("JWT_ACCESS_TTL", 15)) * time.Minute,
			RefreshTTL: time.Duration(getEnvInt("JWT_REFRESH_TTL", 720)) * time.Hour,
		},
		DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "taskmanager"),
		},
		EmailConfig{
			EmailFrom: getEnv("SMTP_FROM_EMAIL", ""),
			EmailPass: getEnv("SMTP_FROM_PASS", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

func (c *Config) GetJWTSecret() string {
	return c.JWT.Secret
}

func (c *Config) GetJWTAccessTTL() time.Duration {
	return c.JWT.AccessTTL
}

func (c *Config) GetJWTRefreshTTL() time.Duration {
	return c.JWT.RefreshTTL
}
func (c *Config) GetDBConnectionString() string {
	return "host=" + c.DB.Host +
		" port=" + strconv.Itoa(c.DB.Port) +
		" user=" + c.DB.User +
		" password=" + c.DB.Password +
		" dbname=" + c.DB.DBName +
		" sslmode=disable"
}

func (c *Config) GetAuthServiceURL() string {
	return c.GetAuthGRPCAddress()
}

func (c *Config) GetTaskServiceURL() string {
	return c.GetTaskGRPCAddress()
}

func (c *Config) GetTaskGRPCAddress() string {
	return getEnv("Task_GRPC_HOST", "localhost") + ":" + strconv.Itoa(getEnvInt("Task_GRPC_PORT", 50052))
}

func (c *Config) GetAuthGRPCAddress() string {
	return getEnv("Auth_GRPC_HOST", "localhost") + ":" + strconv.Itoa(getEnvInt("Auth_GRPC_PORT", 50051))
}
