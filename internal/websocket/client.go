package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"messenger-svyaz/internal/presence"
	"messenger-svyaz/internal/service"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512 * 1024 // 512KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *Hub

	// The websocket connection.
	Conn *websocket.Conn

	// Username
	Username string

	// Buffered channel of outbound messages.
	Send chan []byte

	// Chat rooms this client is in
	ChatRooms map[string]bool

	// Group rooms this client is in
	GroupRooms map[string]bool

	// Rate limiter for this client
	RateLimiter *TokenBucket

	// Rate limiter parent for cleanup
	RateLimiterParent *service.WebSocketRateLimiter

	// Presence manager
	Presence *presence.Presence

	mu sync.Mutex
}

// ReadPump reads messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		if c.Presence != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			c.Presence.SetOffline(ctx, c.Username)
			cancel()
		}
		if c.RateLimiterParent != nil {
			c.RateLimiterParent.Remove(c.Username)
		}
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		if c.Presence != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			c.Presence.Heartbeat(ctx, c.Username)
			cancel()
		}
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Rate limit check
		if c.RateLimiter != nil && !c.RateLimiter.Allow() {
			log.Printf("Rate limit exceeded for user %s", c.Username)
			continue
		}

		// Heartbeat on successful read
		if c.Presence != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			c.Presence.Heartbeat(ctx, c.Username)
			cancel()
		}

		// Handle message
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			for {
				select {
				case msg := <-c.Send:
					w.Write([]byte{'\n'})
					w.Write(msg)
				default:
					goto done
				}
			}
		done:

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming websocket messages
func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	event, ok := msg["event"].(string)
	if !ok {
		log.Printf("Message missing event field")
		return
	}

	switch event {
	case "join_chat":
		c.handleJoinChat(msg)
	case "join_group":
		c.handleJoinGroup(msg)
	case "typing":
		c.handleTyping(msg)
	default:
		log.Printf("Unknown event: %s", event)
	}
}

// handleJoinChat handles join_chat event
func (c *Client) handleJoinChat(msg map[string]interface{}) {
	otherUser, ok := msg["other_user"].(string)
	if !ok {
		return
	}

	// Create room name (sorted usernames to ensure consistency)
	room := getChatRoomName(c.Username, otherUser)

	c.mu.Lock()
	if c.ChatRooms == nil {
		c.ChatRooms = make(map[string]bool)
	}
	c.ChatRooms[room] = true
	c.mu.Unlock()

	c.Hub.AddToChatRoom(room, c.Username)

	// Notify other user if online
	if c.Hub.IsUserOnline(otherUser) {
		response := map[string]interface{}{
			"event": "user_joined_chat",
			"user":  c.Username,
		}
		data, _ := json.Marshal(response)
		c.Hub.SendToUser(otherUser, data)
	}
}

// handleJoinGroup handles join_group event
func (c *Client) handleJoinGroup(msg map[string]interface{}) {
	groupID, ok := msg["group_id"].(string)
	if !ok {
		return
	}

	c.mu.Lock()
	if c.GroupRooms == nil {
		c.GroupRooms = make(map[string]bool)
	}
	c.GroupRooms[groupID] = true
	c.mu.Unlock()

	c.Hub.AddToGroupRoom(groupID, c.Username)
}

// handleTyping handles typing event
func (c *Client) handleTyping(msg map[string]interface{}) {
	otherUser, ok := msg["other_user"].(string)
	if !ok {
		return
	}

	room := getChatRoomName(c.Username, otherUser)

	response := map[string]interface{}{
		"event":   "user_typing",
		"user":    c.Username,
		"room":    room,
		"typing":  true,
	}
	data, _ := json.Marshal(response)

	// Send to other user in the chat
	c.Hub.SendToUser(otherUser, data)
}

// getChatRoomName creates a consistent room name for a chat between two users
func getChatRoomName(user1, user2 string) string {
	if user1 < user2 {
		return user1 + ":" + user2
	}
	return user2 + ":" + user1
}
