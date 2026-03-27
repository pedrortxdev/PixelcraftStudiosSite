package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

type PermissionHandler struct {
	permService *service.PermissionService
}

func NewPermissionHandler(permService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permService: permService}
}

// GetMyPermissions returns permissions for the logged-in user (WITH CONTEXT + UUID)
func (h *PermissionHandler) GetMyPermissions(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse UUID in handler, pass typed value to service
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	perms, err := h.permService.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// HasPermission checks if user has a specific permission (WITH CONTEXT + UUID)
func (h *PermissionHandler) HasPermission(c *gin.Context) {
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

	var req struct {
		Resource models.ResourceType `json:"resource" binding:"required"`
		Action   models.ActionType   `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	hasPerm, err := h.permService.HasPermission(c.Request.Context(), userID, req.Resource, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_permission": hasPerm})
}

// GetAllRolePermissions returns all permissions for all roles (WITH CONTEXT)
func (h *PermissionHandler) GetAllRolePermissions(c *gin.Context) {
	perms, err := h.permService.GetAllRolePermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role permissions"})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// GetRolePermissions returns permissions for a specific role (WITH CONTEXT)
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	role := c.Param("role")

	perms, err := h.permService.GetRolePermissions(c.Request.Context(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role permissions"})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// AddRolePermission adds a permission to a role (WITH CONTEXT)
func (h *PermissionHandler) AddRolePermission(c *gin.Context) {
	role := c.Param("role")

	var req struct {
		Resource models.ResourceType `json:"resource" binding:"required"`
		Action   models.ActionType   `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	if err := h.permService.AddRolePermission(c.Request.Context(), role, req.Resource, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add permission", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission added successfully"})
}

// RemoveRolePermission removes a permission from a role (WITH CONTEXT)
func (h *PermissionHandler) RemoveRolePermission(c *gin.Context) {
	role := c.Param("role")

	var req struct {
		Resource models.ResourceType `json:"resource" binding:"required"`
		Action   models.ActionType   `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	if err := h.permService.RemoveRolePermission(c.Request.Context(), role, req.Resource, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission removed successfully"})
}

// GetAvailableResources returns all available resources
func (h *PermissionHandler) GetAvailableResources(c *gin.Context) {
	resources := []models.ResourceType{
		models.ResourceUsers,
		models.ResourceRoles,
		models.ResourceProducts,
		models.ResourceOrders,
		models.ResourceTransactions,
		models.ResourceSupport,
		models.ResourceEmails,
		models.ResourceFiles,
		models.ResourceGames,
		models.ResourceCategories,
		models.ResourcePlans,
		models.ResourceDashboard,
		models.ResourceSettings,
		models.ResourceSystem,
		models.ResourceDiscounts,
	}

	c.JSON(http.StatusOK, gin.H{"resources": resources})
}

// GetAvailableActions returns all available actions
func (h *PermissionHandler) GetAvailableActions(c *gin.Context) {
	actions := []models.ActionType{
		models.ActionView,
		models.ActionCreate,
		models.ActionEdit,
		models.ActionDelete,
		models.ActionManage,
	}

	c.JSON(http.StatusOK, gin.H{"actions": actions})
}

// GetAvailableRoles returns all available roles in the system
func (h *PermissionHandler) GetAvailableRoles(c *gin.Context) {
	roles := []string{
		"PARTNER",
		"CLIENT",
		"CLIENT_VIP",
		"SUPPORT",
		"ADMIN",
		"DEVELOPMENT",
		"ENGINEERING",
		"DIRECTION",
	}

	roleDescriptions := map[string]string{
		"PARTNER":     "Parceiro: +1% de lucros em vendas",
		"CLIENT":      "Cliente: prioridade 3 estrelas, adquirido com depósito",
		"CLIENT_VIP":  "Cliente VIP: prioridade 4 estrelas, R$200/mês ou assinatura",
		"SUPPORT":     "Suporte: acesso restrito (Atendimento + Email próprio)",
		"ADMIN":       "Administração: visualização total, sem edição",
		"DEVELOPMENT": "Desenvolvimento: edita planos/produtos/jogos/categorias/arquivos",
		"ENGINEERING": "Engenharia: acesso completo exceto cargos",
		"DIRECTION":   "Direção: acesso total incluindo gerenciamento de cargos",
	}

	roleHierarchy := map[string]int{
		"PARTNER":     1,
		"CLIENT":      2,
		"CLIENT_VIP":  3,
		"SUPPORT":     4,
		"ADMIN":       5,
		"DEVELOPMENT": 6,
		"ENGINEERING": 7,
		"DIRECTION":   8,
	}

	result := []map[string]interface{}{}
	for _, role := range roles {
		result = append(result, map[string]interface{}{
			"role":        role,
			"description": roleDescriptions[role],
			"level":       roleHierarchy[role],
			"is_admin":    roleHierarchy[role] >= 4,
		})
	}

	c.JSON(http.StatusOK, gin.H{"roles": result})
}
