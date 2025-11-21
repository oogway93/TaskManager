package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/oogway93/taskmanager/config"
	"github.com/prometheus/client_golang/prometheus"
	// "github.com/oogway93/taskmanager/internal/metrics"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
)

type AuthUser struct {
	UserID   string
	Username string
	Email    string
	Role     string
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
			log.Println("Failed to extract token from header")

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
		c.Set("username", authUser.Username)
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
	username, _ := claims["username"].(string)
	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	if userID == "" || email == "" {
		return nil, fmt.Errorf("invalid user data in token")
	}

	return &AuthUser{
		UserID: userID,
		Username: username,
		Email:  email,
		Role:   role,
	}, nil
}

// isPublicRoute проверяет, является ли маршрут публичным
func isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/health",
		"/api/v1/auth/registration",
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

func PrometheusInit() {
	// Регистрируем метрики
	prometheus.MustRegister(
		HttpRequestsTotal,
		HttpRequestDuration,
		HttpRequestsInFlight,
		AuthRegistrationDuration,
		AuthRegistrations,
		GrpcCallDuration,
		// ordersProcessed,
		// activeUsers,
		DbQueryDuration,
	)
}

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status", "handler"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "path", "status", "handler"},
	)

	HttpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)
// // Бизнес-метрики
	ordersProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_processed_total",
			Help: "Total processed orders",
		},
		[]string{"type", "status"},
	)

	activeUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Current active users",
		},
	)

	// Метрики базы данных
	DbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation", "table"},
	)
    AuthRegistrations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_registrations_total",
            Help: "Total number of user registrations",
        },
        []string{"status"}, // "success", "validation_error", "grpc_error"
    )

    AuthRegistrationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "auth_registration_duration_seconds",
            Help:    "Duration of registration requests",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2},
        },
        []string{"status"},
    )

    GrpcCallDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "grpc_call_duration_seconds",
            Help:    "Duration of gRPC calls",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
        },
        []string{"service", "method"},
    )
)

// Метрики для бизнес-логики
func RecordOrder(orderType, status string) {
	ordersProcessed.WithLabelValues(orderType, status).Inc()
}

func SetActiveUsers(count int) {
	activeUsers.Set(float64(count))
}

func RecordDBQuery(operation, table string, duration time.Duration) {
	DbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// Вспомогательная функция для измерения gRPC вызовов
// func RecordGRPCCall(service, method string, duration time.Duration, err error) {
//     status := "success"
//     if err != nil {
//         status = "error"
//     }
//     GrpcCallDuration.WithLabelValues(service, method).Observe(duration.Seconds())
// }

// PrometheusMiddleware для Gin
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем метрики эндпоинт
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		
		// Если путь не определен (404), используем raw path
		if path == "" {
			path = c.Request.URL.Path
		}

		// Увеличиваем счетчик активных запросов
		HttpRequestsInFlight.Inc()
		defer HttpRequestsInFlight.Dec()

		// Обрабатываем запрос
		c.Next()

		// Получаем статус ответа
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		
		// Регистрируем метрики
		duration := time.Since(start).Seconds()
		HttpRequestsTotal.WithLabelValues(method, path, status, c.HandlerName()).Inc()
		HttpRequestDuration.WithLabelValues(method, path, status, c.HandlerName()).Observe(duration)
	}
}



// ==================== ROLE-BASED ACCESS CONTROL ====================

// // RequireRole middleware для проверки ролей
// func RequireRole(roles ...string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		user, err := GetUserFromContext(c)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "UNAUTHORIZED",
// 				"message": "Authentication required",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Проверяем, есть ли у пользователя необходимая роль
// 		hasRole := false
// 		for _, role := range roles {
// 			if user.Role == role {
// 				hasRole = true
// 				break
// 			}
// 		}

// 		if !hasRole {
// 			// logger.WithFields(logger.Fields{
// 			// 	"user_id":        user.UserID,
// 			// 	"email":          user.Email,
// 			// 	"role":           user.Role,
// 			// 	"required_roles": roles,
// 			// }).Warn("Access denied - insufficient permissions")
// 			log.Println("Access denied - insufficient permissions")

// 			c.JSON(http.StatusForbidden, gin.H{
// 				"error":   "FORBIDDEN",
// 				"message": "Insufficient permissions",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		c.Next()
// 	}
// }

// // AdminOnly middleware для проверки административных прав
// func AdminOnly() gin.HandlerFunc {
// 	return RequireRole("admin", "superadmin")
// }
