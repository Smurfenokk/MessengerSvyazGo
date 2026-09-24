package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/redis"
	"messenger-svyaz/internal/service"
)

type contextKey string

const (
	UserKey  contextKey = "username"
	AdminKey contextKey = "is_admin"
)

type AuthMiddleware struct {
	jwtService *service.JWTService
}

func NewAuthMiddleware(jwtService *service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

// JWT validates JWT token and adds username to context
func (am *AuthMiddleware) JWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"status":"error","message":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"status":"error","message":"Invalid authorization header"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		claims, err := am.jwtService.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"status":"error","message":"Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Add username and admin status to context
		ctx := context.WithValue(r.Context(), UserKey, claims.Username)
		ctx = context.WithValue(ctx, AdminKey, claims.IsAdmin)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUsername extracts username from context
func GetUsername(r *http.Request) string {
	if username, ok := r.Context().Value(UserKey).(string); ok {
		return username
	}
	return ""
}

// IsAdmin checks if user is admin
func IsAdmin(r *http.Request) bool {
	if isAdmin, ok := r.Context().Value(AdminKey).(bool); ok {
		return isAdmin
	}
	return false
}

// RateLimitMiddleware creates a rate limiter middleware using Redis
func RateLimitMiddleware(redisAddr, redisPassword string, rate string) (func(http.Handler) http.Handler, error) {
	// Create a redis store
	store, err := redis.NewStore(redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
	})
	if err != nil {
		return nil, err
	}

	// Create a rate limiter instance
	instance := limiter.Rate{
		Period: limiter.Period(3600 * time.Second), // 1 hour
		Limit:  1000,                               // 1000 requests per hour
	}

	// TODO: parse rate string to configure custom rates per endpoint

	middleware := middleware.NewRateLimiter(store, instance)
	return middleware, nil
}

// MaxBodySizeMiddleware limits request body size
func MaxBodySizeMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ResponseWriter wraps http.ResponseWriter to capture status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// JSONResponse writes a JSON response
func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ErrorResponse writes an error JSON response
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, map[string]interface{}{
		"status":  "error",
		"message": message,
	})
}

// SuccessResponse writes a success JSON response
func SuccessResponse(w http.ResponseWriter, data interface{}) {
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}
