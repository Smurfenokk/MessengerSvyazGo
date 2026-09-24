package websocket

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"messenger-svyaz/internal/handler"
	"messenger-svyaz/internal/service"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub         *Hub
	pubsub      *PubSub
	authHandler *handler.AuthMiddleware
	rateLimiter *service.WebSocketRateLimiter
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *Hub, pubsub *PubSub, authHandler *handler.AuthMiddleware, rateLimiter *service.WebSocketRateLimiter) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		pubsub:      pubsub,
		authHandler: authHandler,
		rateLimiter: rateLimiter,
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

	// Upgrade to WebSocket
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
