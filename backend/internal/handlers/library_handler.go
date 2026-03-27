package handlers

import (
	"fmt"
	"net/http"
	"net/url"

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
// @Success 200 {array} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /library [get]
func (h *LibraryHandler) GetMyLibrary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	items, err := h.service.GetUserLibrary(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve library"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetDownloadURL godoc
// @Summary Download a purchased product
// @Tags library
// @Produce octet-stream
// @Param id path string true "Product ID (UUID)"
// @Success 200 {file} file "Download file"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /library/{id}/download [get]
func (h *LibraryHandler) GetDownloadURL(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	productID := c.Param("id")

	// Validate the product ID is a proper UUID before processing
	_, err := uuid.Parse(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	// Get download info
	downloadPath, isFile, err := h.service.GetDownloadInfo(c.Request.Context(), userID.(string), productID)
	if err != nil {
		if err.Error() == "user does not own this product" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process download request"})
		return
	}

	if isFile {
		// Serve the file directly
		// Get the filename to set proper content-disposition
		parsedProductID, err := uuid.Parse(productID) // Should not fail since we already validated it
		if err != nil {
			// This should not happen but added for safety
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error: invalid product ID"})
			return
		}
		product, err := h.service.GetProduct(c.Request.Context(), parsedProductID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product info"})
			return
		}

		// Set appropriate headers for file download
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		encodedName := url.QueryEscape(product.Name)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", product.Name, encodedName))
		c.Header("Content-Type", "application/octet-stream")

		// Serve the file
		c.File(downloadPath)
	} else {
		// It's a URL-based download, return the URL
		c.JSON(http.StatusOK, gin.H{"download_url": downloadPath})
	}
}