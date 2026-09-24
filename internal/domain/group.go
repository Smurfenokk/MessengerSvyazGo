package domain

import "time"

type Group struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Creator     string                 `json:"creator"`
	CreatedAt   time.Time              `json:"created_at"`
	Settings    map[string]interface{} `json:"settings"`
}

type GroupCreate struct {
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

type GroupJoin struct {
	GroupID string `json:"group_id" validate:"required"`
}

type GroupInfo struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    *string                `json:"description,omitempty"`
	Creator        string                 `json:"creator"`
	Members        []string               `json:"members"`
	MembersCount   int                    `json:"members_count"`
	CreatedAt      string                 `json:"created_at"`
	Settings       map[string]interface{} `json:"settings"`
	IsAdmin        bool                   `json:"is_admin"`
	AllowReactions bool                   `json:"allow_reactions"`
}

type GroupMessage struct {
	ID        string                 `json:"id"`
	GroupID   string                 `json:"group_id"`
	Sender    string                 `json:"sender"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Edited    bool                   `json:"edited"`
	Deleted   bool                   `json:"deleted"`
	ReplyTo   *string                `json:"reply_to,omitempty"`
	Reactions map[string][]string    `json:"reactions,omitempty"`
}

type GroupMessageCreate struct {
	GroupID string `json:"group_id" validate:"required"`
	Content string `json:"message" validate:"required"`
	Type    string `json:"type" validate:"omitempty,oneof=group image file"`
	ReplyTo string `json:"reply_to,omitempty"`
}
