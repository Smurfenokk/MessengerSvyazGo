package domain

import "time"

type SupportMessage struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Answered  bool      `json:"answered"`
}

type SupportMessageCreate struct {
	Content string `json:"message" validate:"required"`
}
