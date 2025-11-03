package repository

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// NewPostgresDB создает подключение к PostgreSQL
func NewPostgresDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}