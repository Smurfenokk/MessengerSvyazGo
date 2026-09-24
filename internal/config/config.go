package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Storage  StorageConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  int
	WriteTimeout int
	MaxBodySize  int64 // 16MB max
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	Expiration int // hours
}

type StorageConfig struct {
	Type string // local or s3
	Path string // for local storage
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// .env is optional in production
	}

	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 15),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 15),
			MaxBodySize:  getEnvInt64("SERVER_MAX_BODY_SIZE", 16*1024*1024), // 16MB
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/messenger?sslmode=disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "change-me-in-production"),
			Expiration: getEnvInt("JWT_EXPIRATION", 24*7), // 7 days
		},
		Storage: StorageConfig{
			Type: getEnv("STORAGE_TYPE", "local"),
			Path: getEnv("STORAGE_PATH", "./data"),
		},
	}

	if cfg.JWT.Secret == "change-me-in-production" {
		fmt.Println("WARNING: Using default JWT secret. Change JWT_SECRET in production!")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultVal
}
