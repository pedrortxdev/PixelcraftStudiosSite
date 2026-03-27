package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

// UserHandler handles user profile HTTP requests
type UserHandler struct {
	userService *service.UserService
	roleService *service.RoleService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService, roleService *service.RoleService) *UserHandler {
	return &UserHandler{
		userService: userService,
		roleService: roleService,
	}
}

// GetProfile handles GET /api/v1/users/me
// @Summary Get user profile
// @Description Retrieves the authenticated user's profile information
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Type assertion
	uid, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// Get user profile
	profile, err := h.userService.GetProfile(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve profile",
			"details": err.Error(),
		})
		return
	}

	// Fetch user roles
	if h.roleService != nil {
		roles, err := h.roleService.GetUserRoles(c.Request.Context(), uid)
		if err == nil {
			profile.Roles = roles
			profile.HighestRole = models.GetHighestRole(roles)
		}
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile handles PUT /api/v1/users/me
// @Summary Update user profile
// @Description Updates the authenticated user's optional profile information
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.UpdateUserRequest true "Profile update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Type assertion
	uid, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req models.UpdateUserRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Update profile
	err := h.userService.UpdateProfile(c.Request.Context(), uid, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update profile",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// UploadAvatar handles POST /api/v1/users/me/avatar
// @Summary Upload and set user avatar
// @Description Uploads an image file and sets it as the user's avatar
// @Tags users
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/me/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userID, ok := userIDStr.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get file from form
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No avatar file provided"})
		return
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit"})
		return
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, 
		".gif": true, ".webp": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: png, jpg, jpeg, gif, webp"})
		return
	}

	// Validate actual content type by reading file header
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	// Read first 512 bytes to detect content type
	header := make([]byte, 512)
	_, err = src.Read(header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Detect content type from magic bytes
	contentType := http.DetectContentType(header)
	allowedMimes := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedMimes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File content does not match allowed image types"})
		return
	}

	// Create directory if not exists
	uploadDir := "./uploads/public/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate safe filename (UUID_Timestamp.ext)
	newFilename := fmt.Sprintf("%s_%d%s", userID, time.Now().Unix(), ext)
	dst := filepath.Join(uploadDir, newFilename)

	// Save file
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar file"})
		return
	}

	// Construct Public URL
	avatarURL := fmt.Sprintf("/public/avatars/%s", newFilename)

	// Update user profile
	req := models.UpdateUserRequest{
		AvatarURL: &avatarURL,
	}
	
	err = h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile with avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar updated successfully",
		"avatar_url": avatarURL,
	})
}