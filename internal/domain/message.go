package domain

import "time"

type Message struct {
	ID         string                 `json:"id"`
	Sender     string                 `json:"sender"`
	Receiver   string                 `json:"receiver"`
	Content    string                 `json:"content"`
	Timestamp  time.Time              `json:"timestamp"`
	Type       string                 `json:"type"`
	Edited     bool                   `json:"edited"`
	Deleted    bool                   `json:"deleted"`
	Read       bool                   `json:"read"`
	ReplyTo    *string                `json:"reply_to,omitempty"`
	Reactions  map[string][]string    `json:"reactions,omitempty"`
}

type MessageCreate struct {
	Receiver string `json:"receiver" validate:"required"`
	Content  string `json:"message" validate:"required"`
	Type     string `json:"type" validate:"omitempty,oneof=text image file saved"`
	ReplyTo  string `json:"reply_to,omitempty"`
}

type MessageUpdate struct {
	Content string `json:"content" validate:"required"`
}

type ReactionAdd struct {
	Emoji string `json:"emoji" validate:"required"`
}

type ChatPreview struct {
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	LastMessage  string `json:"last_message"`
	LastTime     string `json:"last_time"`
	UnreadCount  int    `json:"unread_count"`
	Online       bool   `json:"online"`
}
