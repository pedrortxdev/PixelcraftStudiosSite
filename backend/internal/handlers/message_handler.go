package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// SendMessage handles posting a new message
func (h *MessageHandler) SendMessage(c *gin.Context) {
	subIDStr := c.Param("id")
	userIDStr := c.GetString("user_id") // Key from auth middleware

	isAdmin := c.GetBool("isAdmin")
	if !isAdmin {
		isAdmin = c.GetBool("is_admin")
	}

	// Parse UUIDs at the boundary (Controller responsibility)
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID format"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	var req models.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.messageService.SendMessage(c.Request.Context(), subID, userID, req.Content, isAdmin)
	if err != nil {
		if errors.Is(err, errors.New("unauthorized: you do not own this subscription")) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "message content cannot be empty" || 
		   err.Error() == "message content exceeds maximum length of 10000 characters" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

// GetMessages handles retrieving chat history
func (h *MessageHandler) GetMessages(c *gin.Context) {
	subIDStr := c.Param("id")
	userIDStr := c.GetString("user_id")
	isAdmin := c.GetBool("isAdmin")
	if !isAdmin {
		isAdmin = c.GetBool("is_admin")
	}

	// Parse UUIDs at the boundary (Controller responsibility)
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID format"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Parse pagination parameters (optional)
	limit := service.DefaultChatHistoryLimit
	offset := service.DefaultChatHistoryOffset

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	params := &service.GetChatHistoryParams{
		Limit:  limit,
		Offset: offset,
	}

	messages, err := h.messageService.GetChatHistory(c.Request.Context(), subID, userID, isAdmin, params)
	if err != nil {
		if err.Error() == "unauthorized: you do not own this subscription" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}
