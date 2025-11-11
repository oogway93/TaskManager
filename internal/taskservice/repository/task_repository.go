package repository

import (
	"database/sql"
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type TaskRepository interface {
	// Create(ctx context.Context, user *entity.User) error
	// GetByID(ctx context.Context, userID string) (*entity.User, error)
	// GetByEmail(ctx context.Context, email string) (*entity.User, error)
	// // Update(ctx context.Context, user *service.User) error
	// ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository создает новый репозиторий пользователей
func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

// func (r *taskRepository) Create(ctx context.Context, user *entity.User) error {
// 	query := `
// 		INSERT INTO users (id, email, password_hash, name, role, active, created_at, updated_at)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
// 	` //TODO: убрать raw sql, использовать gORM

// 	user.ID = uuid.New().String()
// 	user.CreatedAt = time.Now()
// 	user.UpdatedAt = time.Now()

// 	_, err := r.db.ExecContext(ctx, query,
// 		user.ID,
// 		user.Email,
// 		user.Password, // уже захэшированный пароль
// 		user.Name,
// 		user.Role,
// 		user.Active,
// 		user.CreatedAt,
// 		user.UpdatedAt,
// 	)

// 	return err
// }

// // ExistsByEmail проверяет существование пользователя с email
// func (r *taskRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
// 	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
// 	var exists bool
// 	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
// 	return exists, err
// }

// func (r *taskRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
// 	query := `
// 		SELECT id, email, password_hash, name, role, active, created_at, updated_at
// 		FROM users 
// 		WHERE email = $1
// 	`

// 	var user entity.User
// 	err := r.db.QueryRowContext(ctx, query, email).Scan(
// 		&user.ID,
// 		&user.Email,
// 		&user.Password,
// 		&user.Name,
// 		&user.Role,
// 		&user.Active,
// 		&user.CreatedAt,
// 		&user.UpdatedAt,
// 	)

// 	if err == sql.ErrNoRows {
// 		return nil, ErrUserNotFound
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &user, nil
// }

// func (r *taskRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
// 	query := `
// 		SELECT id, email, password_hash, name, role, active, created_at, updated_at
// 		FROM users 
// 		WHERE id = $1 AND active = true
// 	`

// 	var user entity.User
// 	err := r.db.QueryRowContext(ctx, query, userID).Scan(
// 		&user.ID,
// 		&user.Email,
// 		&user.Password,
// 		&user.Name,
// 		&user.Role,
// 		&user.Active,
// 		&user.CreatedAt,
// 		&user.UpdatedAt,
// 	)

// 	if err == sql.ErrNoRows {
// 		return nil, ErrUserNotFound
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &user, nil
// }
