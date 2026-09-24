package handler

import (
	"encoding/json"
	"net/http"

	"messenger-svyaz/internal/domain"
	"messenger-svyaz/internal/id"
	"messenger-svyaz/internal/repository"
)

type SupportHandler struct {
	supportRepo *repository.SupportRepository
}

func NewSupportHandler(supportRepo *repository.SupportRepository) *SupportHandler {
	return &SupportHandler{
		supportRepo: supportRepo,
	}
}

// GetSupportMessages returns support messages for current user
func (h *SupportHandler) GetSupportMessages(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	messages, err := h.supportRepo.GetByUser(r.Context(), currentUser)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get support messages")
		return
	}

	SuccessResponse(w, map[string]interface{}{
		"messages": messages,
	})
}

// SendSupportMessage sends a message to support
func (h *SupportHandler) SendSupportMessage(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req domain.SupportMessageCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Generate message ID
	messageID := id.GenerateShortID()

	// Create support message
	err := h.supportRepo.Create(r.Context(), messageID, currentUser, currentUser, req.Content, "user")
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	SuccessResponse(w, nil)
}
