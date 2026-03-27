package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/service"
)

// LibraryHandler handles HTTP requests for user library and downloads
type LibraryHandler struct {
	service *service.LibraryService
}

func NewLibraryHandler(service *service.LibraryService) *LibraryHandler {
	return &LibraryHandler{service: service}
}

// GetMyLibrary godoc
// @Summary Get user's purchased products
// @Tags library
// @Produce json
// @Success 200 {array} service.LibraryItem
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /library [get]
func (h *LibraryHandler) GetMyLibrary(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Handler parses UUID, Service receives typed uuid.UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	items, err := h.service.GetUserLibrary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve library"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetDownloadURL godoc
// @Summary Get one-time download URL for a product
// @Tags library
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} service.DownloadInfo
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /library/{id}/download [get]
func (h *LibraryHandler) GetDownloadURL(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	// Get ONE-TIME download URL (not static!)
	info, err := h.service.GetDownloadInfo(c.Request.Context(), userID, productID)
	if err != nil {
		if err.Error() == "user does not own this product" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't own this product"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get download info"})
		return
	}

	// Return ephemeral URL with token (expires in 15 min, 1 use only)
	c.JSON(http.StatusOK, gin.H{
		"download_url":  info.DownloadURL,
		"expires_at":    info.ExpiresAt,
		"max_downloads": info.MaxDownloads,
		"is_one_time":   info.IsOneTime,
		"warning":       "This link expires after one use or 15 minutes. Do not share.",
	})
}

// ValidateOneTimeDownload validates and processes a one-time download token
// @Summary Validate one-time download token and download file
// @Tags library
// @Produce octet-stream
// @Param token path string true "One-time download token (UUID)"
// @Success 200 {file} file "Download file"
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /files/onedownload/{token} [get]
func (h *LibraryHandler) ValidateOneTimeDownload(c *gin.Context) {
	tokenStr := c.Param("token")
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token format"})
		return
	}

	// Get client info for logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Validate token and get file info
	file, err := h.service.ValidateDownloadToken(c.Request.Context(), token, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// File is valid - send it
	c.FileAttachment(file.FilePath, file.Name)
}
