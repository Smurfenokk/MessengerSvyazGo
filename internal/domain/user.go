package domain

import "time"

type User struct {
	Username     string    `json:"username"`
	AccountID    string    `json:"account_id"`
	PasswordHash string    `json:"-"` // never expose in JSON
	FirstName    string    `json:"first_name"`
	LastName     *string   `json:"last_name,omitempty"`
	MiddleName   *string   `json:"middle_name,omitempty"`
	DisplayName  string    `json:"display_name"`
	IsAdmin      bool      `json:"is_admin"`
	IsOnline     bool      `json:"is_online"`
	IsBlocked    bool      `json:"is_blocked"`
	Bio          string    `json:"bio"`
	AvatarKey    *string   `json:"avatar_key,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
}

type UserCreate struct {
	Username     string  `json:"username" validate:"required,min=3,max=32,lowercase"`
	Password     string  `json:"password" validate:"required,min=8"`
	FirstName    string  `json:"first_name" validate:"required"`
	LastName     string  `json:"last_name,omitempty"`
	MiddleName   string  `json:"middle_name,omitempty"`
	DisplayName  string  `json:"display_name" validate:"required"`
	RegistrationKey string `json:"key" validate:"required"`
}

type UserLogin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserUpdate struct {
	Bio string `json:"bio" validate:"max=30"`
}

type PasswordChange struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type UserPublic struct {
	Username    string  `json:"username"`
	AccountID   string  `json:"account_id"`
	DisplayName string  `json:"display_name"`
	IsAdmin     bool    `json:"is_admin"`
	IsOnline    bool    `json:"is_online"`
	IsBlocked   bool    `json:"is_blocked"`
	Bio         string  `json:"bio"`
	HasAvatar   bool    `json:"has_avatar"`
	IsVisible   bool    `json:"is_visible"`
}
