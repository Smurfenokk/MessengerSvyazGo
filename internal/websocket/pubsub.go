package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

// PubSub handles Redis Pub/Sub for multi-instance WebSocket communication
type PubSub struct {
	client *redis.Client
	hub    *Hub
	ctx    context.Context
}

// NewPubSub creates a new PubSub instance
func NewPubSub(client *redis.Client, hub *Hub) *PubSub {
	return &PubSub{
		client: client,
		hub:    hub,
		ctx:    context.Background(),
	}
}

// Subscribe subscribes to Redis channels and forwards messages to the hub
func (ps *PubSub) Subscribe(channels ...string) error {
	pubsub := ps.client.Subscribe(ps.ctx, channels...)
	_, err := pubsub.Receive(ps.ctx)
	if err != nil {
		return err
	}

	ch := pubsub.Channel()

	go func() {
		for msg := range ch {
			ps.handleMessage(msg.Payload)
		}
	}()

	return nil
}

// Publish publishes a message to a Redis channel
func (ps *PubSub) Publish(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return ps.client.Publish(ps.ctx, channel, data).Err()
}

// handleMessage processes messages from Redis Pub/Sub
func (ps *PubSub) handleMessage(payload string) {
	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		log.Printf("Invalid Pub/Sub message: %v", err)
		return
	}

	event, ok := msg["event"].(string)
	if !ok {
		return
	}

	switch event {
	case "new_message":
		ps.handleNewMessage(msg)
	case "new_group_message":
		ps.handleNewGroupMessage(msg)
	case "user_typing":
		ps.handleUserTyping(msg)
	case "user_joined_chat":
		ps.handleUserJoinedChat(msg)
	}
}

// handleNewMessage handles new message from Pub/Sub
func (ps *PubSub) handleNewMessage(msg map[string]interface{}) {
	sender, _ := msg["sender"].(string)
	receiver, _ := msg["receiver"].(string)

	// Send to receiver if they're connected to this instance
	if ps.hub.IsUserOnline(receiver) {
		data, _ := json.Marshal(msg)
		ps.hub.SendToUser(receiver, data)
	}
}

// handleNewGroupMessage handles new group message from Pub/Sub
func (ps *PubSub) handleNewGroupMessage(msg map[string]interface{}) {
	groupID, _ := msg["group_id"].(string)

	// Send to all users in the group room
	data, _ := json.Marshal(msg)
	ps.hub.SendToGroupRoom(groupID, data)
}

// handleUserTyping handles typing indicator from Pub/Sub
func (ps *PubSub) handleUserTyping(msg map[string]interface{}) {
	otherUser, _ := msg["other_user"].(string)

	// Send to the other user if they're connected to this instance
	if ps.hub.IsUserOnline(otherUser) {
		data, _ := json.Marshal(msg)
		ps.hub.SendToUser(otherUser, data)
	}
}

// handleUserJoinedChat handles user joined chat from Pub/Sub
func (ps *PubSub) handleUserJoinedChat(msg map[string]interface{}) {
	otherUser, _ := msg["other_user"].(string)

	// Send to the other user if they're connected to this instance
	if ps.hub.IsUserOnline(otherUser) {
		data, _ := json.Marshal(msg)
		ps.hub.SendToUser(otherUser, data)
	}
}

// PublishNewMessage publishes a new message to Pub/Sub
func (ps *PubSub) PublishNewMessage(sender, receiver string, message map[string]interface{}) error {
	msg := map[string]interface{}{
		"event":    "new_message",
		"sender":   sender,
		"receiver": receiver,
		"message":  message,
	}
	return ps.Publish("messages", msg)
}

// PublishNewGroupMessage publishes a new group message to Pub/Sub
func (ps *PubSub) PublishNewGroupMessage(groupID, sender string, message map[string]interface{}) error {
	msg := map[string]interface{}{
		"event":    "new_group_message",
		"group_id": groupID,
		"sender":   sender,
		"message":  message,
	}
	return ps.Publish("group_messages", msg)
}

// PublishUserTyping publishes a typing indicator to Pub/Sub
func (ps *PubSub) PublishUserTyping(sender, otherUser string) error {
	msg := map[string]interface{}{
		"event":      "user_typing",
		"user":       sender,
		"other_user": otherUser,
		"typing":     true,
	}
	return ps.Publish("typing", msg)
}

// PublishUserJoinedChat publishes a user joined chat event to Pub/Sub
func (ps *PubSub) PublishUserJoinedChat(user, otherUser string) error {
	msg := map[string]interface{}{
		"event":      "user_joined_chat",
		"user":       user,
		"other_user": otherUser,
	}
	return ps.Publish("chat_events", msg)
}
