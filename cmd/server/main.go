package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"messenger-svyaz/internal/config"
	"messenger-svyaz/internal/db"
	"messenger-svyaz/internal/db/sqlc"
	"messenger-svyaz/internal/handler"
	"messenger-svyaz/internal/repository"
	"messenger-svyaz/internal/service"
	"messenger-svyaz/internal/websocket"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Initialize database
	database, err := db.New(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize sqlc queries
	queries := sqlc.New(database.Pool)

	// Initialize repositories
	userRepo := repository.NewUserRepository(queries)
	keyRepo := repository.NewKeyRepository(queries)
	messageRepo := repository.NewMessageRepository(queries)
	groupRepo := repository.NewGroupRepository(queries)
	supportRepo := repository.NewSupportRepository(queries)

	// Initialize services
	pwService := &service.PasswordService{}
	jwtService := service.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiration)
	wsRateLimiter := service.NewWebSocketRateLimiter()

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize WebSocket Pub/Sub
	pubsub := websocket.NewPubSub(redisClient, hub)
	if err := pubsub.Subscribe("messages", "group_messages", "typing", "chat_events"); err != nil {
		log.Fatalf("Failed to subscribe to Redis Pub/Sub: %v", err)
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userRepo, keyRepo, jwtService, pwService)
	userHandler := handler.NewUserHandler(userRepo, pwService)
	chatHandler := handler.NewChatHandler(messageRepo)
	groupHandler := handler.NewGroupHandler(groupRepo, userRepo)
	supportHandler := handler.NewSupportHandler(supportRepo)

	// Initialize middleware
	authMiddleware := handler.NewAuthMiddleware(jwtService)

	// Initialize WebSocket handler
	wsHandler := websocket.NewWebSocketHandler(hub, pubsub, authMiddleware, wsRateLimiter)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(time.Duration(cfg.Server.ReadTimeout)*time.Second))
	r.Use(handler.CORSMiddleware)
	r.Use(handler.MaxBodySizeMiddleware(cfg.Server.MaxBodySize))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Health(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(fmt.Sprintf(`{"status":"error","message":"database unreachable: %v"}`, err)))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth endpoints (no JWT required)
	r.Post("/login", authHandler.Login)
	r.Post("/register", authHandler.Register)

	// WebSocket endpoint
	r.Mount("/ws", wsHandler.Routes())

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		r.Use(authMiddleware.JWT)

		// User endpoints
		r.Get("/users/list", userHandler.GetUsersList)
		r.Get("/profile", userHandler.GetProfile)
		r.Post("/profile/update", userHandler.UpdateProfile)
		r.Post("/profile/change_password", userHandler.ChangePassword)
		r.Get("/avatar/{username}", userHandler.GetAvatar)
		r.Post("/upload/avatar", userHandler.UploadAvatar)

		// Chat endpoints
		r.Get("/chat/messages/{other_user}", chatHandler.GetChatMessages)
		r.Get("/chat/messages/saved", chatHandler.GetSavedMessages)
		r.Post("/chat/send", chatHandler.SendMessage)
		r.Post("/chat/save", chatHandler.SaveMessage)
		r.Get("/chat/recent", chatHandler.GetRecentChats)

		// Group endpoints
		r.Get("/groups/my", groupHandler.GetMyGroups)
		r.Post("/groups/create", groupHandler.CreateGroup)
		r.Post("/groups/join", groupHandler.JoinGroup)
		r.Get("/group/messages/{group_id}", groupHandler.GetGroupMessages)
		r.Get("/group/info/{group_id}", groupHandler.GetGroupInfo)
		r.Post("/group/send", groupHandler.SendGroupMessage)
		r.Post("/group/reaction", groupHandler.AddGroupReaction)

		// Support endpoints
		r.Get("/support/messages", supportHandler.GetSupportMessages)
		r.Post("/support/send", supportHandler.SendSupportMessage)
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
