package postgres

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// NewPostgresDB создает подключение к PostgreSQL
func NewPostgresDB(connStr string, Log *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		Log.Fatal("Failed connection to postgres DB in Auth Service")
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Проверка подключения
	if err := db.Ping(); err != nil {
		Log.Fatal("Failed to ping db connection:", zap.Error(err))
		return nil, err
	}

	return db, nil
}