package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"messenger-svyaz/internal/domain"
	"messenger-svyaz/internal/id"
	"messenger-svyaz/internal/repository"
)

type ChatHandler struct {
	messageRepo *repository.MessageRepository
}

func NewChatHandler(messageRepo *repository.MessageRepository) *ChatHandler {
	return &ChatHandler{
		messageRepo: messageRepo,
	}
}

// GetChatMessages returns chat history with another user
func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	otherUser := chi.URLParam(r, "other_user")
	if otherUser == "" {
		ErrorResponse(w, http.StatusBadRequest, "Other user required")
		return
	}

	messages, err := h.messageRepo.GetChatMessages(r.Context(), currentUser, otherUser)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	SuccessResponse(w, map[string]interface{}{
		"messages": messages,
	})
}

// GetSavedMessages returns saved messages
func (h *ChatHandler) GetSavedMessages(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	messages, err := h.messageRepo.GetSavedMessages(r.Context(), currentUser)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get saved messages")
		return
	}

	SuccessResponse(w, map[string]interface{}{
		"messages": messages,
	})
}

// SendMessage sends a message to another user
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req domain.MessageCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Generate message ID
	messageID := id.GenerateShortID()

	// Set default type if not provided
	if req.Type == "" {
		req.Type = "text"
	}

	// Create message
	err := h.messageRepo.Create(r.Context(), messageID, currentUser, req.Receiver, req.Content, req.Type, nullableString(req.ReplyTo))
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	// TODO: emit via WebSocket to receiver

	SuccessResponse(w, map[string]interface{}{
		"message_id": messageID,
	})
}

// SaveMessage saves a message to saved messages
func (h *ChatHandler) SaveMessage(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Generate message ID
	messageID := id.GenerateShortID()

	// Create saved message (sender = receiver = current user)
	err := h.messageRepo.Create(r.Context(), messageID, currentUser, currentUser, req.Message, "saved", nil)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to save message")
		return
	}

	SuccessResponse(w, map[string]interface{}{
		"message_id": messageID,
	})
}

// GetRecentChats returns recent chats
func (h *ChatHandler) GetRecentChats(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chats, err := h.messageRepo.GetRecentChats(r.Context(), currentUser)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get recent chats")
		return
	}

	// TODO: build proper chat preview with user info
	// For now return raw data
	SuccessResponse(w, map[string]interface{}{
		"chats": chats,
	})
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
