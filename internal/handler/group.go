package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"messenger-svyaz/internal/domain"
	"messenger-svyaz/internal/id"
	"messenger-svyaz/internal/repository"
)

type GroupHandler struct {
	groupRepo *repository.GroupRepository
	userRepo  *repository.UserRepository
}

func NewGroupHandler(groupRepo *repository.GroupRepository, userRepo *repository.UserRepository) *GroupHandler {
	return &GroupHandler{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// GetMyGroups returns groups where user is a member or creator
func (h *GroupHandler) GetMyGroups(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	groups, err := h.groupRepo.GetWithMember(r.Context(), currentUser)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get groups")
		return
	}

	result := make([]map[string]interface{}, 0)
	for _, g := range groups {
		members, _ := h.groupRepo.GetMembers(r.Context(), g.ID)
		role := "member"
		if g.Creator == currentUser {
			role = "creator"
		}

		// Extract allow_reactions from settings
		allowReactions := true
		if settings, ok := g.Settings.(map[string]interface{}); ok {
			if ar, ok := settings["allow_reactions"].(bool); ok {
				allowReactions = ar
			}
		}

		result = append(result, map[string]interface{}{
			"id":             g.ID,
			"name":           g.Name,
			"description":    g.Description,
			"creator":        g.Creator,
			"members_count":  len(members),
			"role":           role,
			"settings": map[string]interface{}{
				"allow_reactions": allowReactions,
			},
		})
	}

	SuccessResponse(w, map[string]interface{}{
		"groups": result,
	})
}

// CreateGroup creates a new group
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req domain.GroupCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set default settings if not provided
	if req.Settings == nil {
		req.Settings = map[string]interface{}{"allow_reactions": true}
	}

	// Generate group ID
	groupID := id.GenerateShortID()

	// Create group
	err := h.groupRepo.Create(r.Context(), groupID, req.Name, req.Description, currentUser, req.Settings)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	// Add creator as member
	h.groupRepo.AddMember(r.Context(), groupID, currentUser)

	SuccessResponse(w, map[string]interface{}{
		"group_id": groupID,
	})
}

// JoinGroup joins a group
func (h *GroupHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req domain.GroupJoin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if group exists
	group, err := h.groupRepo.Get(r.Context(), req.GroupID)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "Group not found")
		return
	}

	// Add member
	err = h.groupRepo.AddMember(r.Context(), req.GroupID, currentUser)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to join group")
		return
	}

	SuccessResponse(w, nil)
}

// GetGroupMessages returns messages from a group
func (h *GroupHandler) GetGroupMessages(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	groupID := chi.URLParam(r, "group_id")
	if groupID == "" {
		ErrorResponse(w, http.StatusBadRequest, "Group ID required")
		return
	}

	// Check if user is member or creator
	group, err := h.groupRepo.Get(r.Context(), groupID)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "Group not found")
		return
	}

	members, _ := h.groupRepo.GetMembers(r.Context(), groupID)
	isMember := false
	for _, m := range members {
		if m == currentUser {
			isMember = true
			break
		}
	}

	if !isMember && group.Creator != currentUser {
		ErrorResponse(w, http.StatusForbidden, "Access denied")
		return
	}

	// Get messages
	messages, err := h.groupRepo.GetMessages(r.Context(), groupID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	// Extract allow_reactions from settings
	allowReactions := true
	if settings, ok := group.Settings.(map[string]interface{}); ok {
		if ar, ok := settings["allow_reactions"].(bool); ok {
			allowReactions = ar
		}
	}

	SuccessResponse(w, map[string]interface{}{
		"messages":         messages,
		"is_admin":         group.Creator == currentUser,
		"allow_reactions":  allowReactions,
		"members":          members,
		"creator":          group.Creator,
	})
}

// GetGroupInfo returns group information
func (h *GroupHandler) GetGroupInfo(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	groupID := chi.URLParam(r, "group_id")
	if groupID == "" {
		ErrorResponse(w, http.StatusBadRequest, "Group ID required")
		return
	}

	group, err := h.groupRepo.Get(r.Context(), groupID)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "Group not found")
		return
	}

	members, _ := h.groupRepo.GetMembers(r.Context(), groupID)
	isMember := false
	for _, m := range members {
		if m == currentUser {
			isMember = true
			break
		}
	}

	if !isMember && group.Creator != currentUser {
		ErrorResponse(w, http.StatusForbidden, "Access denied")
		return
	}

	SuccessResponse(w, map[string]interface{}{
		"id":            group.ID,
		"name":          group.Name,
		"description":   group.Description,
		"creator":       group.Creator,
		"members":       members,
		"members_count": len(members),
		"created_at":    group.CreatedAt,
		"settings":      group.Settings,
	})
}

// SendGroupMessage sends a message to a group
func (h *GroupHandler) SendGroupMessage(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req domain.GroupMessageCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set default type if not provided
	if req.Type == "" {
		req.Type = "group"
	}

	// Generate message ID
	messageID := id.GenerateShortID()

	// Create message
	err := h.groupRepo.CreateMessage(r.Context(), messageID, req.GroupID, currentUser, req.Content, req.Type, nullableString(req.ReplyTo))
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	// TODO: emit via WebSocket to group

	SuccessResponse(w, nil)
}

// AddGroupReaction adds a reaction to a group message
func (h *GroupHandler) AddGroupReaction(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUsername(r)
	if currentUser == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		GroupID   string `json:"group_id"`
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: implement reaction logic
	// For now return success
	SuccessResponse(w, nil)
}
