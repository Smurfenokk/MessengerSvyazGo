package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"messenger-svyaz/internal/repository"
	"messenger-svyaz/internal/service"
)

type UserHandler struct {
	userRepo   *repository.UserRepository
	pwService  *service.PasswordService
}

func NewUserHandler(userRepo *repository.UserRepository, pwService *service.PasswordService) *UserHandler {
	return &UserHandler{
		userRepo:  userRepo,
		pwService: pwService,
	}
}

// GetUsersList returns list of users
func (h *UserHandler) GetUsersList(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	excludeAdmins := r.URL.Query().Get("exclude_admins") == "1"
	search := r.URL.Query().Get("search")

	users, err := h.userRepo.List(r.Context(), excludeAdmins, search)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		hasAvatar := u.AvatarKey.Valid
		result = append(result, map[string]interface{}{
			"username":     u.Username,
			"account_id":   u.AccountID,
			"display_name": u.DisplayName,
			"is_admin":     u.IsAdmin,
			"is_online":    u.IsOnline,
			"is_blocked":   u.IsBlocked,
			"bio":          u.Bio,
			"has_avatar":   hasAvatar,
			"is_visible":   true,
		})
	}

	SuccessResponse(w, map[string]interface{}{
		"users": result,
	})
}

// GetProfile returns current user profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userRepo.GetByUsername(r.Context(), currentUser)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	avatarURL := ""
	if user.AvatarKey.Valid {
		avatarURL = "/api/avatar/" + user.Username
	}

	SuccessResponse(w, map[string]interface{}{
		"username":     user.Username,
		"display_name": user.DisplayName,
		"bio":          user.Bio,
		"has_avatar":   user.AvatarKey.Valid,
		"avatar_url":   avatarURL,
	})
}

// UpdateProfile updates user bio
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Bio string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Limit bio to 30 characters
	if len(req.Bio) > 30 {
		req.Bio = req.Bio[:30]
	}

	err := h.userRepo.UpdateBio(r.Context(), currentUser, req.Bio)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	SuccessResponse(w, nil)
}

// ChangePassword changes user password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user
	user, err := h.userRepo.GetByUsername(r.Context(), currentUser)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Verify current password
	if !h.pwService.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		ErrorResponse(w, http.StatusBadRequest, "Invalid current password")
		return
	}

	// Validate new password
	if len(req.NewPassword) < 8 {
		ErrorResponse(w, http.StatusBadRequest, "New password must be at least 8 characters")
		return
	}

	// Hash new password
	newHash, err := h.pwService.HashPassword(req.NewPassword)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Update password
	err = h.userRepo.UpdatePassword(r.Context(), currentUser, newHash)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	SuccessResponse(w, nil)
}

// GetAvatar returns user avatar
func (h *UserHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		ErrorResponse(w, http.StatusBadRequest, "Username required")
		return
	}

	user, err := h.userRepo.GetByUsername(r.Context(), username)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	if !user.AvatarKey.Valid {
		ErrorResponse(w, http.StatusNotFound, "Avatar not found")
		return
	}

	// TODO: implement avatar retrieval from storage
	// For now, return error
	ErrorResponse(w, http.StatusNotImplemented, "Avatar storage not implemented yet")
}

// UploadAvatar uploads user avatar
func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// TODO: implement file upload
	ErrorResponse(w, http.StatusNotImplemented, "Avatar upload not implemented yet")
}
