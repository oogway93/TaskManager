package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	App App
}

type App struct {
	Env string
}

type ServerConfig struct {
	Host string
	Port int
}

func Load() *Config {
	return &Config{
		ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnvInt("SERVER_PORT", 8000),
		},
		App{
			Env: getEnv("APP_ENV", "development"),
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