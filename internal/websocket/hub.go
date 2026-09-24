package websocket

import (
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients by username
	clients map[string]*Client

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Map of chat rooms (chat usernames) to clients
	chatRooms map[string]map[string]bool // room -> username -> true

	// Map of group rooms (group IDs) to clients
	groupRooms map[string]map[string]bool // group_id -> username -> true

	mu sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		chatRooms:  make(map[string]map[string]bool),
		groupRooms: make(map[string]map[string]bool),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.Username] = client
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.Username]; ok {
				delete(h.clients, client.Username)
				// Remove from all rooms
				for room := range client.ChatRooms {
					h.removeFromChatRoom(room, client.Username)
				}
				for room := range client.GroupRooms {
					h.removeFromGroupRoom(room, client.Username)
				}
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			deadClients := make([]*Client, 0)
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Client send buffer is full, close connection
					deadClients = append(deadClients, client)
				}
			}
			h.mu.RUnlock()

			// Remove dead clients with write lock
			h.mu.Lock()
			for _, client := range deadClients {
				close(client.Send)
				delete(h.clients, client.Username)
			}
			h.mu.Unlock()
		}
	}
}

// AddToChatRoom adds a user to a chat room
func (h *Hub) AddToChatRoom(room, username string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.chatRooms[room]; !ok {
		h.chatRooms[room] = make(map[string]bool)
	}
	h.chatRooms[room][username] = true
}

// RemoveFromChatRoom removes a user from a chat room
func (h *Hub) RemoveFromChatRoom(room, username string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeFromChatRoom(room, username)
}

func (h *Hub) removeFromChatRoom(room, username string) {
	if roomUsers, ok := h.chatRooms[room]; ok {
		delete(roomUsers, username)
		if len(roomUsers) == 0 {
			delete(h.chatRooms, room)
		}
	}
}

// AddToGroupRoom adds a user to a group room
func (h *Hub) AddToGroupRoom(room, username string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.groupRooms[room]; !ok {
		h.groupRooms[room] = make(map[string]bool)
	}
	h.groupRooms[room][username] = true
}

// RemoveFromGroupRoom removes a user from a group room
func (h *Hub) RemoveFromGroupRoom(room, username string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeFromGroupRoom(room, username)
}

func (h *Hub) removeFromGroupRoom(room, username string) {
	if roomUsers, ok := h.groupRooms[room]; ok {
		delete(roomUsers, username)
		if len(roomUsers) == 0 {
			delete(h.groupRooms, room)
		}
	}
}

// GetChatRoomUsers returns users in a chat room
func (h *Hub) GetChatRoomUsers(room string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0)
	if roomUsers, ok := h.chatRooms[room]; ok {
		for username := range roomUsers {
			users = append(users, username)
		}
	}
	return users
}

// GetGroupRoomUsers returns users in a group room
func (h *Hub) GetGroupRoomUsers(room string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0)
	if roomUsers, ok := h.groupRooms[room]; ok {
		for username := range roomUsers {
			users = append(users, username)
		}
	}
	return users
}

// SendToChatRoom sends a message to all users in a chat room
func (h *Hub) SendToChatRoom(room string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomUsers, ok := h.chatRooms[room]; ok {
		for username := range roomUsers {
			if client, ok := h.clients[username]; ok {
				select {
				case client.Send <- message:
				default:
					// Client send buffer is full
				}
			}
		}
	}
}

// SendToGroupRoom sends a message to all users in a group room
func (h *Hub) SendToGroupRoom(room string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomUsers, ok := h.groupRooms[room]; ok {
		for username := range roomUsers {
			if client, ok := h.clients[username]; ok {
				select {
				case client.Send <- message:
				default:
					// Client send buffer is full
				}
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(username string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.clients[username]; ok {
		select {
		case client.Send <- message:
		default:
			// Client send buffer is full
		}
	}
}

// IsUserOnline checks if a user is online
func (h *Hub) IsUserOnline(username string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, ok := h.clients[username]
	return ok
}

// GetOnlineCount returns the number of online users
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}
