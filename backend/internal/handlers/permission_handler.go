package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

type PermissionHandler struct {
	permService *service.PermissionService
}

func NewPermissionHandler(permService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permService: permService}
}

// GetMyPermissions retorna as permissões do usuário logado
func (h *PermissionHandler) GetMyPermissions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	perms, err := h.permService.GetUserPermissions(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// GetAllRolePermissions retorna todas as permissões de todos os cargos
func (h *PermissionHandler) GetAllRolePermissions(c *gin.Context) {
	perms, err := h.permService.GetAllRolePermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role permissions"})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// GetRolePermissions retorna as permissões de um cargo específico
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	role := c.Param("role")

	perms, err := h.permService.GetRolePermissions(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role permissions"})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// AddRolePermission adiciona uma permissão a um cargo
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

	if err := h.permService.AddRolePermission(role, req.Resource, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add permission", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission added successfully"})
}

// RemoveRolePermission remove uma permissão de um cargo
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

	if err := h.permService.RemoveRolePermission(role, req.Resource, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission removed successfully"})
}

// GetAvailableResources retorna todos os recursos disponíveis
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

// GetAvailableActions retorna todas as ações disponíveis
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

// GetAvailableRoles retorna todos os cargos disponíveis no sistema
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
		"CLIENT_VIP": "Cliente VIP: prioridade 4 estrelas, R$200/mês ou assinatura",
		"SUPPORT":     "Suporte: acesso restrito (Atendimento + Email próprio)",
		"ADMIN":       "Administração: visualização total, sem edição",
		"DEVELOPMENT": "Desenvolvimento: edita planos/produtos/jogos/categorias/arquivos",
		"ENGINEERING": "Engenharia: acesso completo exceto cargos",
		"DIRECTION":   "Direção: acesso total incluindo gerenciamento de cargos",
	}

	roleHierarchy := map[string]int{
		"PARTNER":    1,
		"CLIENT":     2,
		"CLIENT_VIP": 3,
		"SUPPORT":    4,
		"ADMIN":      5,
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
