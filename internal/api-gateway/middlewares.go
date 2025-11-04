package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/oogway93/taskmanager/config"
)

type AuthUser struct {
	UserID string
	Email  string
	Role   string
}

// JWTConfig конфигурация для JWT middleware
type JWTConfig struct {
	SecretKey string
}

// NewJWTConfig создает конфигурацию JWT
func NewJWTConfig(cfg *config.Config) *JWTConfig {
	return &JWTConfig{
		SecretKey: cfg.JWT.Secret,
	}
}

func JWTMiddleware(jwtConfig *JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем публичные маршруты
		if isPublicRoute(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Извлекаем токен из заголовка
		tokenString, err := extractTokenFromHeader(c)
		if err != nil {
			// logger.WithFields(logger.Fields{
			// 	"path":   c.Request.URL.Path,
			// 	"method": c.Request.Method,
			// 	"error":  err.Error(),
			// }).Warn("Failed to extract token from header")
			log.Println("Failed to extract token from header" )
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Authorization token required",
			})
			c.Abort()
			return
		}

		// Проверяем токен локально
		authUser, err := validateTokenLocally(tokenString, jwtConfig.SecretKey)
		if err != nil {
			// logger.WithFields(logger.Fields{
			// 	"path":   c.Request.URL.Path,
			// 	"method": c.Request.Method,
			// 	"error":  err.Error(),
			// }).Warn("Token validation failed")
			log.Println("Token validation failed")
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "INVALID_TOKEN",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Устанавливаем данные пользователя в контекст
		c.Set("user", authUser)
		c.Set("user_id", authUser.UserID)
		c.Set("user_email", authUser.Email)
		c.Set("user_role", authUser.Role)

		// logger.WithFields(logger.Fields{
		// 	"user_id": authUser.UserID,
		// 	"email":   authUser.Email,
		// 	"role":    authUser.Role,
		// 	"path":    c.Request.URL.Path,
		// }).Debug("User authenticated")
		log.Println("User authentificated")

		c.Next()
	}
}

// extractTokenFromHeader извлекает JWT токен из заголовка Authorization
func extractTokenFromHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	// Проверяем формат: Bearer <token>
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization format, expected: Bearer <token>")
	}

	return parts[1], nil
}

// validateTokenLocally проверяет токен локально без вызова auth service
func validateTokenLocally(tokenString, secretKey string) (*AuthUser, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Извлекаем claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Проверяем expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}

	// Извлекаем данные пользователя
	userID, _ := claims["user_id"].(string)
	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	if userID == "" || email == "" {
		return nil, fmt.Errorf("invalid user data in token")
	}

	return &AuthUser{
		UserID: userID,
		Email:  email,
		Role:   role,
	}, nil
}

// isPublicRoute проверяет, является ли маршрут публичным
func isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/health",
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/metrics",
	}

	for _, route := range publicRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}

	return false
}

// GetUserFromContext извлекает пользователя из контекста
func GetUserFromContext(c *gin.Context) (*AuthUser, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, fmt.Errorf("user not found in context")
	}

	authUser, ok := user.(*AuthUser)
	if !ok {
		return nil, fmt.Errorf("invalid user type in context")
	}

	return authUser, nil
}

// ==================== ROLE-BASED ACCESS CONTROL ====================

// RequireRole middleware для проверки ролей
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// Проверяем, есть ли у пользователя необходимая роль
		hasRole := false
		for _, role := range roles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			// logger.WithFields(logger.Fields{
			// 	"user_id":        user.UserID,
			// 	"email":          user.Email,
			// 	"role":           user.Role,
			// 	"required_roles": roles,
			// }).Warn("Access denied - insufficient permissions")
			log.Println("Access denied - insufficient permissions")
			
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "FORBIDDEN",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly middleware для проверки административных прав
func AdminOnly() gin.HandlerFunc {
	return RequireRole("admin", "superadmin")
}