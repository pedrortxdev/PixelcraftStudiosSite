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

// GetPermissionAuditLog retorna o histórico de mudanças de permissões
func (h *PermissionAdvancedHandler) GetPermissionAuditLog(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 100 {
		limit = 100 // BT-031
	}
	role := c.Query("role")

	logs, total, err := h.service.GetPermissionAuditLog(page, limit, role)
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

// InheritPermissions herda permissões de um cargo inferior
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

	count, err := h.service.InheritPermissions(targetRole, req.SourceRole, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to inherit permissions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Permissions inherited successfully",
		"inherited_count":  count,
		"target_role":      targetRole,
		"source_role":      req.SourceRole,
	})
}

// RemoveInheritedPermissions remove permissões herdadas de um cargo
func (h *PermissionAdvancedHandler) RemoveInheritedPermissions(c *gin.Context) {
	role := c.Param("role")

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	count, err := h.service.RemoveInheritedPermissions(role, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove inherited permissions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Inherited permissions removed successfully",
		"removed_count":  count,
		"role":           role,
	})
}

// CreateCustomRole cria um novo cargo customizado
func (h *PermissionAdvancedHandler) CreateCustomRole(c *gin.Context) {
	var req struct {
		RoleName      string `json:"role_name" binding:"required"`
		DisplayName   string `json:"display_name" binding:"required"`
		Description   string `json:"description"`
		Color         string `json:"color"`
		HierarchyLevel int   `json:"hierarchy_level" binding:"required,min=1,max=10"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	role, err := h.service.CreateCustomRole(req.RoleName, req.DisplayName, req.Description, req.Color, req.HierarchyLevel, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom role", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// GetCustomRoles retorna todos os cargos customizados
func (h *PermissionAdvancedHandler) GetCustomRoles(c *gin.Context) {
	roles, err := h.service.GetCustomRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get custom roles", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// DeleteCustomRole deleta um cargo customizado
func (h *PermissionAdvancedHandler) DeleteCustomRole(c *gin.Context) {
	roleID := c.Param("id")

	if err := h.service.DeleteCustomRole(roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete custom role", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Custom role deleted successfully"})
}

// ExportPermissions exporta configurações de permissões
func (h *PermissionAdvancedHandler) ExportPermissions(c *gin.Context) {
	roles := c.QueryArray("roles")

	data, err := h.service.ExportPermissions(roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export permissions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// ImportPermissions importa configurações de permissões
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

	result, err := h.service.ImportPermissions(req.TemplateData, req.Overwrite, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import permissions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SavePermissionTemplate salva um template de permissões
func (h *PermissionAdvancedHandler) SavePermissionTemplate(c *gin.Context) {
	var req struct {
		TemplateName string                 `json:"template_name" binding:"required"`
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

	template, err := h.service.SavePermissionTemplate(req.TemplateName, req.Description, req.TemplateData, req.IsPublic, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save template", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetPermissionTemplates retorna todos os templates disponíveis
func (h *PermissionAdvancedHandler) GetPermissionTemplates(c *gin.Context) {
	templates, err := h.service.GetPermissionTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// GetPermissionNotifications retorna notificações de permissões do usuário
func (h *PermissionAdvancedHandler) GetPermissionNotifications(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	notifications, err := h.service.GetUserNotifications(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

// MarkNotificationAsRead marca uma notificação como lida
func (h *PermissionAdvancedHandler) MarkNotificationAsRead(c *gin.Context) {
	notificationID := c.Param("id")

	if err := h.service.MarkNotificationAsRead(notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// GetPermissionDashboard retorna estatísticas do dashboard de auditoria
func (h *PermissionAdvancedHandler) GetPermissionDashboard(c *gin.Context) {
	stats, err := h.service.GetPermissionDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard stats", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
