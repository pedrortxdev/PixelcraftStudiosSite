package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/service"
)

type PermissionAdvancedHandler struct {
	service *service.PermissionAdvancedService
}

func NewPermissionAdvancedHandler(service *service.PermissionAdvancedService) *PermissionAdvancedHandler {
	return &PermissionAdvancedHandler{service: service}
}

// GetPermissionAuditLog returns permission change audit log
func (h *PermissionAdvancedHandler) GetPermissionAuditLog(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 100 {
		limit = 100
	}
	role := c.Query("role")

	logs, total, err := h.service.GetPermissionAuditLog(c.Request.Context(), page, limit, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit log", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// InheritPermissions inherits permissions from a source role
func (h *PermissionAdvancedHandler) InheritPermissions(c *gin.Context) {
	targetRole := c.Param("role")

	var req struct {
		SourceRole string `json:"source_role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	count, err := h.service.InheritPermissions(c.Request.Context(), targetRole, req.SourceRole, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to inherit permissions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Permissions inherited successfully",
		"inherited_count": count,
		"target_role":     targetRole,
		"source_role":     req.SourceRole,
	})
}

// RemoveInheritedPermissions removes inherited permissions from a role
func (h *PermissionAdvancedHandler) RemoveInheritedPermissions(c *gin.Context) {
	role := c.Param("role")

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	count, err := h.service.RemoveInheritedPermissions(c.Request.Context(), role, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove inherited permissions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Inherited permissions removed successfully",
		"removed_count": count,
		"role":          role,
	})
}

// CreateCustomRole creates a new custom role
func (h *PermissionAdvancedHandler) CreateCustomRole(c *gin.Context) {
	var req struct {
		RoleName       string `json:"role_name" binding:"required"`
		DisplayName    string `json:"display_name" binding:"required"`
		Description    string `json:"description"`
		Color          string `json:"color"`
		HierarchyLevel int    `json:"hierarchy_level" binding:"required,min=1,max=10"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	role, err := h.service.CreateCustomRole(c.Request.Context(), req.RoleName, req.DisplayName, req.Description, req.Color, req.HierarchyLevel, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom role", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// GetCustomRoles returns all custom roles
func (h *PermissionAdvancedHandler) GetCustomRoles(c *gin.Context) {
	roles, err := h.service.GetCustomRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get custom roles", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// DeleteCustomRole deactivates a custom role
func (h *PermissionAdvancedHandler) DeleteCustomRole(c *gin.Context) {
	roleID := c.Param("id")

	err := h.service.DeleteCustomRole(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete custom role", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Custom role deleted successfully"})
}

// ExportPermissions exports permission configurations
func (h *PermissionAdvancedHandler) ExportPermissions(c *gin.Context) {
	var req struct {
		Roles []string `json:"roles"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	data, err := h.service.ExportPermissions(c.Request.Context(), req.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export permissions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// ImportPermissions imports permission configurations
func (h *PermissionAdvancedHandler) ImportPermissions(c *gin.Context) {
	var req struct {
		TemplateData map[string]interface{} `json:"template_data" binding:"required"`
		Overwrite    bool                   `json:"overwrite"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	result, err := h.service.ImportPermissions(c.Request.Context(), req.TemplateData, req.Overwrite, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import permissions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SavePermissionTemplate saves a permission template
func (h *PermissionAdvancedHandler) SavePermissionTemplate(c *gin.Context) {
	var req struct {
		Name         string                 `json:"name" binding:"required"`
		Description  string                 `json:"description"`
		TemplateData map[string]interface{} `json:"template_data" binding:"required"`
		IsPublic     bool                   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	template, err := h.service.SavePermissionTemplate(c.Request.Context(), req.Name, req.Description, req.TemplateData, req.IsPublic, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save template", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetPermissionTemplates returns all public templates
func (h *PermissionAdvancedHandler) GetPermissionTemplates(c *gin.Context) {
	templates, err := h.service.GetPermissionTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// GetUserNotifications returns notifications for the current user
func (h *PermissionAdvancedHandler) GetUserNotifications(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	notifications, err := h.service.GetUserNotifications(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

// MarkNotificationAsRead marks a notification as read
func (h *PermissionAdvancedHandler) MarkNotificationAsRead(c *gin.Context) {
	notificationID := c.Param("id")

	err := h.service.MarkNotificationAsRead(c.Request.Context(), notificationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// GetPermissionDashboard returns permission statistics
func (h *PermissionAdvancedHandler) GetPermissionDashboard(c *gin.Context) {
	stats, err := h.service.GetPermissionDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard stats", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// BulkImportPermissions performs a bulk import of permissions
func (h *PermissionAdvancedHandler) BulkImportPermissions(c *gin.Context) {
	var req struct {
		Permissions []service.PermissionImport `json:"permissions" binding:"required"`
		Overwrite   bool                       `json:"overwrite"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	count, err := h.service.BulkImportPermissions(c.Request.Context(), req.Permissions, req.Overwrite)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk import permissions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk import completed",
		"count":   count,
	})
}
