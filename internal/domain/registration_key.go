package domain

import "time"

type RegistrationKey struct {
	Key      string     `json:"key"`
	CreatedAt time.Time  `json:"created_at"`
	IsActive bool       `json:"is_active"`
	UsedBy   *string    `json:"used_by,omitempty"`
	UsedAt   *time.Time `json:"used_at,omitempty"`
}
