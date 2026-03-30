package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

type AIHandler struct {
	aiService   *service.AIService
	userService *service.UserService
}

func NewAIHandler(aiService *service.AIService, userService *service.UserService) *AIHandler {
	return &AIHandler{
		aiService:   aiService,
		userService: userService,
	}
}

type FormatTextRequest struct {
	Text string `json:"text" binding:"required"`
}

func (h *AIHandler) FormatText(c *gin.Context) {
	var req FormatTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use context with timeout for AI operations
	ctx, cancel := c.Request.Context(), func() {}
	defer cancel()

	formattedText, err := h.aiService.FormatText(ctx, req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to format text: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"formatted_text": formattedText,
	})
}

type GenerateAvatarRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	UserID string `json:"user_id"` // Optional, if provided by admin for another user
}

func (h *AIHandler) GenerateAvatar(c *gin.Context) {
	var req GenerateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user ID from context
	targetUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid := targetUserID.(string)

	// If requesting to generate for a different user, verify admin permission
	if req.UserID != "" && req.UserID != uid {
		// Check if current user is admin
		isAdmin, adminExists := c.Get("is_admin")
		if !adminExists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can generate avatars for other users"})
			return
		}
		uid = req.UserID
	}

	// Generate Image via AI with context for timeout/cancellation
	ctx, cancel := c.Request.Context(), func() {}
	defer cancel()

	base64Data, err := h.aiService.GenerateAvatar(ctx, req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate AI avatar: " + err.Error()})
		return
	}

	// Decode Base64
	imgBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode generated image"})
		return
	}

	// Save to uploads/public/avatars
	uploadDir := "./uploads/public/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	filename := fmt.Sprintf("ai_%s_%d.png", uid, time.Now().Unix())
	dst := filepath.Join(uploadDir, filename)

	if err := os.WriteFile(dst, imgBytes, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar file"})
		return
	}

	// Update user profile
	avatarURL := fmt.Sprintf("/public/avatars/%s", filename)
	updateReq := models.UpdateUserRequest{
		AvatarURL: &avatarURL,
	}

	err = h.userService.UpdateProfile(c.Request.Context(), uid, &updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "AI Avatar generated and updated successfully",
		"avatar_url": avatarURL,
	})
}
