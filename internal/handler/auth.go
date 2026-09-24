package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"messenger-svyaz/internal/domain"
	"messenger-svyaz/internal/id"
	"messenger-svyaz/internal/repository"
	"messenger-svyaz/internal/service"
)

type AuthHandler struct {
	userRepo    *repository.UserRepository
	keyRepo     *repository.KeyRepository
	jwtService  *service.JWTService
	pwService   *service.PasswordService
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	keyRepo *repository.KeyRepository,
	jwtService *service.JWTService,
	pwService *service.PasswordService,
) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		keyRepo:    keyRepo,
		jwtService: jwtService,
		pwService:  pwService,
	}
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByUsername(r.Context(), req.Username)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Check password
	if !h.pwService.CheckPassword(req.Password, user.PasswordHash) {
		ErrorResponse(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Check if blocked
	if user.IsBlocked {
		ErrorResponse(w, http.StatusForbidden, "Account is blocked")
		return
	}

	// Update last login
	h.userRepo.UpdateLastLogin(r.Context(), req.Username)

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.Username, user.IsAdmin)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	SuccessResponse(w, map[string]interface{}{
		"is_admin":    user.IsAdmin,
		"session_id":  token,
		"username":    user.Username,
		"display_name": user.DisplayName,
	})
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.UserCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate registration key
	key, err := h.keyRepo.Get(r.Context(), req.RegistrationKey)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid or used registration key")
		return
	}

	if !key.IsActive || key.UsedBy != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid or used registration key")
		return
	}

	// Check if username already exists
	_, err = h.userRepo.GetByUsername(r.Context(), req.Username)
	if err == nil {
		ErrorResponse(w, http.StatusConflict, "Username already taken")
		return
	}

	// Validate username length
	if len(req.Username) < 3 || len(req.Username) > 32 {
		ErrorResponse(w, http.StatusBadRequest, "Username must be between 3 and 32 characters")
		return
	}

	// Validate password length
	if len(req.Password) < 8 {
		ErrorResponse(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Validate first name
	if req.FirstName == "" {
		ErrorResponse(w, http.StatusBadRequest, "First name is required")
		return
	}

	// Set display name to first name if not provided
	if req.DisplayName == "" {
		req.DisplayName = req.FirstName
	}

	// Hash password
	passwordHash, err := h.pwService.HashPassword(req.Password)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Generate account ID
	accountID := id.GenerateShortID()

	// Create user
	err = h.userRepo.Create(r.Context(), req.Username, accountID, passwordHash, req.FirstName, req.LastName, req.MiddleName, req.DisplayName, false, "")
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Mark registration key as used
	h.keyRepo.Use(r.Context(), req.Username, req.RegistrationKey)

	SuccessResponse(w, map[string]interface{}{
		"message": "Registration successful",
	})
}
