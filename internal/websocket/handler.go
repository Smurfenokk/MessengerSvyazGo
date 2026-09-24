package websocket

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"messenger-svyaz/internal/handler"
	"messenger-svyaz/internal/presence"
	"messenger-svyaz/internal/service"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub             *Hub
	pubsub          *PubSub
	authHandler     *handler.AuthMiddleware
	rateLimiter     *service.WebSocketRateLimiter
	allowedOrigins  []string
	presence        *presence.Presence
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *Hub, pubsub *PubSub, authHandler *handler.AuthMiddleware, rateLimiter *service.WebSocketRateLimiter, allowedOrigins []string, presence *presence.Presence) *WebSocketHandler {
	return &WebSocketHandler{
		hub:            hub,
		pubsub:         pubsub,
		authHandler:    authHandler,
		rateLimiter:    rateLimiter,
		allowedOrigins: allowedOrigins,
		presence:       presence,
	}
}

// ServeWebSocket handles WebSocket upgrade
func (wh *WebSocketHandler) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract username from JWT token
	username := handler.GetUsername(r)
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade to WebSocket with origin check
	origin := r.Header.Get("Origin")
	allowed := false
	for _, allowedOrigin := range wh.allowedOrigins {
		if origin == allowedOrigin {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client
	client := &Client{
		Hub:         wh.hub,
		Conn:        conn,
		Username:    username,
		Send:        make(chan []byte, 256),
		ChatRooms:   make(map[string]bool),
		GroupRooms:  make(map[string]bool),
		RateLimiter: wh.rateLimiter.GetOrCreate(username),
		Presence:    wh.presence,
	}

	// Mark user as online
	if wh.presence != nil {
		if err := wh.presence.SetOnline(context.Background(), username); err != nil {
			log.Printf("Failed to set user online: %v", err)
		}
	}

	// Register client
	wh.hub.register <- client

	// Start goroutines
	go client.WritePump()
	go client.ReadPump()
}

// Routes returns WebSocket routes
func (wh *WebSocketHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(wh.authHandler.JWT)
	r.Get("/", wh.ServeWebSocket)
	return r
}
