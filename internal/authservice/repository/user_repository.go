package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/oogway93/taskmanager/internal/entity"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type userRepository struct {
	db  *sql.DB
	Log *zap.Logger
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *sql.DB, Log *zap.Logger) UserRepository {
	return &userRepository{
		db:  db,
		Log: Log,
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, username, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
	` //TODO: убрать raw sql, использовать gORM

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Password, // уже захэшированный пароль
		user.Username,
		user.Role,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// ExistsByEmail проверяет существование пользователя с email
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, username, role, active, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	var user entity.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Username,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		r.Log.Error("SQL error 'ErrNoRows' caused in repo's GetByEmail", zap.Error(err)) 
		return nil, ErrUserNotFound
	}
	if err != nil {
		r.Log.Error("Error caused in repo's GetByEmail", zap.Error(err)) 
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, username, role, active, created_at, updated_at
		FROM users 
		WHERE id = $1 AND active = true
	`

	var user entity.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Username,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		r.Log.Error("SQL error 'ErrNoRows' caused in repo's GetByID", zap.Error(err)) 
		return nil, ErrUserNotFound
	}
	if err != nil {
		r.Log.Error("Error caused in repo's GetByID", zap.Error(err)) 
		return nil, err
	}

	return &user, nil
}
