package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

// FileHandler handles HTTP requests for file management
type FileHandler struct {
	service *service.FileService
}

// NewFileHandler creates a new FileHandler
func NewFileHandler(service *service.FileService) *FileHandler {
	return &FileHandler{service: service}
}

// UploadFile godoc
// @Summary Upload a new file
// @Description Upload a file with a friendly name
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param name formData string true "Friendly name for the file"
// @Success 201 {object} models.FileResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File name is required"})
		return
	}

	ctx := c.Request.Context()
	savedFile, err := h.service.SaveFile(ctx, file, userID, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file: %v", err)})
		return
	}

	response := models.FileResponse{
		ID:        savedFile.ID,
		Name:      savedFile.Name,
		FileName:  savedFile.FileName,
		FileType:  savedFile.FileType,
		Size:      savedFile.Size,
		Url:       fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/%s/download", savedFile.ID),
		CreatedAt: savedFile.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// ListFiles godoc
// @Summary List user files
// @Description Get a paginated list of files uploaded by the user
// @Tags files
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.ListFilesResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files [get]
func (h *FileHandler) ListFiles(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	ctx := c.Request.Context()
	files, total, err := h.service.GetFilesByUserID(ctx, userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve files"})
		return
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + pageSize - 1) / pageSize)
	}

	var fileResponses []models.FileResponse
	for _, file := range files {
		response := models.FileResponse{
			ID:        file.ID,
			Name:      file.Name,
			FileName:  file.FileName,
			FileType:  file.FileType,
			Size:      file.Size,
			Url:       fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/%s/download", file.ID),
			CreatedAt: file.CreatedAt,
		}
		fileResponses = append(fileResponses, response)
	}

	response := models.ListFilesResponse{
		Files:      fileResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetFile godoc
// @Summary Get file details
// @Description Get details about a specific file
// @Tags files
// @Produce json
// @Param id path string true "File ID (UUID)"
// @Success 200 {object} models.FileResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /files/{id} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	ctx := c.Request.Context()
	file, err := h.service.GetFileByID(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file"})
		return
	}
	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if file.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	response := models.FileResponse{
		ID:        file.ID,
		Name:      file.Name,
		FileName:  file.FileName,
		FileType:  file.FileType,
		Size:      file.Size,
		Url:       fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/%s/download", file.ID),
		CreatedAt: file.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateFile godoc
// @Summary Update file name
// @Description Update the friendly name of a file
// @Tags files
// @Accept json
// @Produce json
// @Param id path string true "File ID (UUID)"
// @Param request body models.UpdateFileRequest true "Update file data"
// @Success 200 {object} models.FileResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/{id} [put]
func (h *FileHandler) UpdateFile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var req models.UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	err = h.service.UpdateFileName(ctx, fileID, userID, *req.Name)
	if err != nil {
		if err.Error() == "file not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file"})
		return
	}

	updatedFile, err := h.service.GetFileByID(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated file"})
		return
	}

	response := models.FileResponse{
		ID:        updatedFile.ID,
		Name:      updatedFile.Name,
		FileName:  updatedFile.FileName,
		FileType:  updatedFile.FileType,
		Size:      updatedFile.Size,
		Url:       fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/%s/download", updatedFile.ID),
		CreatedAt: updatedFile.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteFile godoc
// @Summary Delete a file
// @Description Soft delete a file from the system
// @Tags files
// @Produce json
// @Param id path string true "File ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	ctx := c.Request.Context()
	err = h.service.DeleteFile(ctx, fileID, userID)
	if err != nil {
		if err.Error() == "file not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// GetFilesForProductSelection retrieves files that can be assigned to products
func (h *FileHandler) GetFilesForProductSelection(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	ctx := c.Request.Context()
	files, total, err := h.service.GetFilesByUserID(ctx, userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve files"})
		return
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + pageSize - 1) / pageSize)
	}

	var fileResponses []models.FileResponse
	for _, file := range files {
		response := models.FileResponse{
			ID:        file.ID,
			Name:      file.Name,
			FileName:  file.FileName,
			FileType:  file.FileType,
			Size:      file.Size,
			Url:       fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/%s/download", file.ID),
			CreatedAt: file.CreatedAt,
		}
		fileResponses = append(fileResponses, response)
	}

	response := models.ListFilesResponse{
		Files:      fileResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// DownloadFile handles file downloads
func (h *FileHandler) DownloadFile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	isAdmin, _ := c.Get("is_admin")
	isAdminBool := false
	if isAdmin != nil {
		if val, ok := isAdmin.(bool); ok {
			isAdminBool = val
		}
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	ctx := c.Request.Context()
	file, err := h.service.GetFileForDownload(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file"})
		return
	}
	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if file.CreatedBy != userID && !isAdminBool {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	filePath := h.service.GetFilePath(file.ID, file.FileName)

	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File not found on disk"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")

	// Get extension from internal file name
	ext := filepath.Ext(file.FileName)
	
	// Use friendly name for download, ensure it has the correct extension
	downloadName := file.Name
	if ext != "" {
		// Remove any existing extension from the friendly name
		nameWithoutExt := strings.TrimSuffix(downloadName, filepath.Ext(downloadName))
		downloadName = nameWithoutExt + ext
	} else {
		// If no extension, use the friendly name as is
		downloadName = file.Name
	}

	// RFC 5987 percent-encoding for international filenames
	encodedName := strings.ReplaceAll(url.QueryEscape(downloadName), "+", "%20")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", downloadName, encodedName))
	c.Header("Content-Type", "application/octet-stream")

	c.File(filePath)
}

// ListAllFiles returns all files for admin view
// GET /api/v1/admin/files
func (h *FileHandler) ListAllFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	ctx := c.Request.Context()
	files, total, err := h.service.GetAllFiles(ctx, page, pageSize, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve files"})
		return
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	var fileResponses []models.FileResponseWithUser
	for _, file := range files {
		response := models.FileResponseWithUser{
			ID:        file.ID,
			Name:      file.Name,
			FileName:  file.FileName,
			FileType:  file.FileType,
			Size:      file.Size,
			Url:       fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/%s/download", file.ID),
			CreatedAt: file.CreatedAt,
			CreatedBy: file.CreatedBy,
			UserEmail: file.UserEmail,
			UserName:  file.UserName,
		}
		fileResponses = append(fileResponses, response)
	}

	response := models.ListFilesAdminResponse{
		Files:      fileResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetFilePermissions retrieves the permission configuration for a file
// GET /api/v1/files/:id/permissions
func (h *FileHandler) GetFilePermissions(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	ctx := c.Request.Context()
	permissions, err := h.service.GetFilePermissions(ctx, fileID, userID)
	if err != nil {
		if err.Error() == "file not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve permissions"})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// UpdateFilePermissions updates the access permissions for a file
// PUT /api/v1/files/:id/permissions
func (h *FileHandler) UpdateFilePermissions(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var req models.UpdateFilePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	updatedFile, err := h.service.UpdateFilePermissions(ctx, fileID, userID, req)
	if err != nil {
		if err.Error() == "file not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permissions"})
		return
	}

	var allowedRoles []string
	if updatedFile.AllowedRoles != nil && len(updatedFile.AllowedRoles) > 0 {
		json.Unmarshal(updatedFile.AllowedRoles, &allowedRoles)
	}

	var allowedProductIDs []string
	if updatedFile.AllowedProductIDs != nil && len(updatedFile.AllowedProductIDs) > 0 {
		json.Unmarshal(updatedFile.AllowedProductIDs, &allowedProductIDs)
	}

	response := models.FilePermission{
		FileID:              updatedFile.ID,
		AccessType:          updatedFile.AccessType,
		RequiredRole:        updatedFile.RequiredRole,
		AllowedRoles:        allowedRoles,
		RequiredProductID:   updatedFile.RequiredProductID,
		AllowedProductIDs:   allowedProductIDs,
		PublicLinkToken:     updatedFile.PublicLinkToken,
		PublicLinkExpiresAt: updatedFile.PublicLinkExpiresAt,
		DownloadCount:       updatedFile.DownloadCount,
		MaxDownloads:        updatedFile.MaxDownloads,
		PublicLinkURL:       fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/public/%s/download", updatedFile.PublicLinkToken.String()),
	}

	c.JSON(http.StatusOK, response)
}

// GetFileAccessLogs retrieves access logs for a file
// GET /api/v1/files/:id/access-logs
func (h *FileHandler) GetFileAccessLogs(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	isAdmin, _ := c.Get("is_admin")
	isAdminBool := isAdmin != nil && isAdmin.(bool)

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	ctx := c.Request.Context()
	file, err := h.service.GetFileByID(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file"})
		return
	}
	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if file.CreatedBy != userID && !isAdminBool {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	logs, total, err := h.service.GetFileAccessLogs(ctx, fileID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	response := models.ListFileAccessLogsResponse{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// AddRolePermission adds a role permission to a file
// POST /api/v1/files/:id/permissions/roles
func (h *FileHandler) AddRolePermission(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	err = h.service.AddRoleToFilePermission(ctx, fileID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add role permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role permission added successfully"})
}

// RemoveRolePermission removes a role permission from a file
// DELETE /api/v1/files/:id/permissions/roles/:role
func (h *FileHandler) RemoveRolePermission(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	role := c.Param("role")

	ctx := c.Request.Context()
	err = h.service.RemoveRoleFromFilePermission(ctx, fileID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role permission removed successfully"})
}

// AddProductPermission adds a product permission to a file
// POST /api/v1/files/:id/permissions/products
func (h *FileHandler) AddProductPermission(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var req struct {
		ProductID uuid.UUID `json:"product_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	err = h.service.AddProductToFilePermission(ctx, fileID, req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product permission added successfully"})
}

// RemoveProductPermission removes a product permission from a file
// DELETE /api/v1/files/:id/permissions/products/:product_id
func (h *FileHandler) RemoveProductPermission(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	productIDStr := c.Param("product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	ctx := c.Request.Context()
	err = h.service.RemoveProductFromFilePermission(ctx, fileID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove product permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product permission removed successfully"})
}

// RegeneratePublicLink regenerates the public link token for a file
// POST /api/v1/files/:id/regenerate-public-link
func (h *FileHandler) RegeneratePublicLink(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	ctx := c.Request.Context()
	publicLinkURL, err := h.service.RegeneratePublicLink(ctx, fileID, userID)
	if err != nil {
		if err.Error() == "file not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate public link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"public_link_url": publicLinkURL})
}

// DownloadFilePublic handles file downloads using public token
// GET /api/v1/files/public/:token/download
func (h *FileHandler) DownloadFilePublic(c *gin.Context) {
	// Parse token
	tokenParam := c.Param("token")
	token, err := uuid.Parse(tokenParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	ctx := c.Request.Context()

	// Get file by token from database
	var file models.File
	query := `
		SELECT id, name, file_name, file_type, file_path, size,
		       created_by, created_at, updated_at, is_deleted,
		       access_type, required_role, allowed_roles, required_product_id, allowed_product_ids,
		       public_link_token, public_link_expires_at, download_count, max_downloads
		FROM files
		WHERE public_link_token = $1 AND is_deleted = false
	`

	err = h.service.GetDB().QueryRowContext(ctx, query, token).Scan(
		&file.ID, &file.Name, &file.FileName, &file.FileType, &file.FilePath,
		&file.Size, &file.CreatedBy, &file.CreatedAt, &file.UpdatedAt, &file.IsDeleted,
		&file.AccessType, &file.RequiredRole, &file.AllowedRoles, &file.RequiredProductID,
		&file.AllowedProductIDs, &file.PublicLinkToken, &file.PublicLinkExpiresAt,
		&file.DownloadCount, &file.MaxDownloads,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found or invalid token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file"})
		return
	}

	// Check if public link is expired
	if file.PublicLinkExpiresAt != nil && file.PublicLinkExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Public link has expired"})
		return
	}

	// Check max downloads
	if file.MaxDownloads != nil && file.DownloadCount >= *file.MaxDownloads {
		c.JSON(http.StatusForbidden, gin.H{"error": "Download limit reached"})
		return
	}

	// Get file path
	filePath := h.service.GetFilePath(file.ID, file.FileName)

	// Check if file exists on disk
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File not found on disk"})
		return
	}

	// Log access (async, don't wait)
	go func() {
		h.service.LogFileAccess(ctx, file.ID, file.CreatedBy, "DOWNLOAD", true, "Public link access", "", "")
	}()

	// Set appropriate headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")

	// Get extension from internal file name
	ext := filepath.Ext(file.FileName)
	
	// Use friendly name for download, ensure it has the correct extension
	downloadName := file.Name
	if ext != "" {
		// Remove any existing extension from the friendly name
		nameWithoutExt := strings.TrimSuffix(downloadName, filepath.Ext(downloadName))
		downloadName = nameWithoutExt + ext
	} else {
		// If no extension, use the friendly name as is
		downloadName = file.Name
	}

	// RFC 5987 percent-encoding for international filenames
	encodedName := strings.ReplaceAll(url.QueryEscape(downloadName), "+", "%20")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", downloadName, encodedName))
	c.Header("Content-Type", "application/octet-stream")

	// Serve the file
	c.File(filePath)
}

// GenerateOneTimeDownloadLink generates a one-time download link for a private file
// POST /api/v1/files/:id/generate-one-time-link
func (h *FileHandler) GenerateOneTimeDownloadLink(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var req struct {
		ExpiresInMinutes int `json:"expires_in_minutes"`
		MaxDownloads     int `json:"max_downloads"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Use defaults if no body provided
		req.ExpiresInMinutes = 15
		req.MaxDownloads = 1
	}

	ctx := c.Request.Context()

	// Generate token
	token, err := h.service.GenerateOneTimeDownloadToken(ctx, fileID, userID, req.ExpiresInMinutes, req.MaxDownloads)
	if err != nil {
		if err.Error() == "file not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		if err.Error() == "unauthorized: file does not belong to user" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate token: %v", err)})
		return
	}

	// Get file name for response
	file, err := h.service.GetFileByID(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file details"})
		return
	}

	// Generate download URL
	downloadURL := fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/one-time/%s/download", token.Token.String())

	response := models.OneTimeDownloadLinkResponse{
		TokenID:      token.ID,
		FileID:       fileID,
		FileName:     file.Name,
		DownloadURL:  downloadURL,
		ExpiresAt:    token.ExpiresAt,
		MaxDownloads: token.MaxDownloads,
		CreatedAt:    token.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DownloadFileOneTime handles file downloads using one-time token
// GET /api/v1/files/one-time/:token/download
func (h *FileHandler) DownloadFileOneTime(c *gin.Context) {
	// Parse token
	tokenParam := c.Param("token")
	token, err := uuid.Parse(tokenParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token format"})
		return
	}

	ctx := c.Request.Context()

	// Get IP and user agent for logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Validate and use token (this will increment download count)
	fileID, isValid, errorMessage, err := h.service.ValidateOneTimeDownloadToken(ctx, token, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
		return
	}

	if !isValid {
		c.JSON(http.StatusForbidden, gin.H{"error": errorMessage})
		return
	}

	// Get file details
	file, err := h.service.GetFileByID(ctx, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file"})
		return
	}
	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Get file path
	filePath := h.service.GetFilePath(file.ID, file.FileName)

	// Check if file exists on disk
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File not found on disk"})
		return
	}

	// Log access (async, don't wait)
	go func() {
		h.service.LogFileAccess(ctx, file.ID, file.CreatedBy, "DOWNLOAD", true, "One-time token access", ipAddress, userAgent)
	}()

	// Set appropriate headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")

	// Get extension from internal file name
	ext := filepath.Ext(file.FileName)

	// Use friendly name for download, ensure it has the correct extension
	downloadName := file.Name
	if ext != "" {
		// Remove any existing extension from the friendly name
		nameWithoutExt := strings.TrimSuffix(downloadName, filepath.Ext(downloadName))
		downloadName = nameWithoutExt + ext
	} else {
		// If no extension, use the friendly name as is
		downloadName = file.Name
	}

	// RFC 5987 percent-encoding for international filenames
	encodedName := strings.ReplaceAll(url.QueryEscape(downloadName), "+", "%20")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", downloadName, encodedName))
	c.Header("Content-Type", "application/octet-stream")

	// Serve the file
	c.File(filePath)
}
