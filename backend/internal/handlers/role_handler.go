package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

// RoleHandler handles role management HTTP requests
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// AddRoleRequest represents the request body for adding a role
type AddRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// AddUserRole adds a role to a user
// POST /admin/users/:id/roles
func (h *RoleHandler) AddUserRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req AddRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate role type
	role := models.RoleType(req.Role)
	if !role.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role type"})
		return
	}

	// Get admin user ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminIDStr, ok := adminID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin ID"})
		return
	}

	// Hierarchy check: admin can only modify users with lower hierarchy (unless legacy admin)
	isLegacyAdmin, legacyErr := h.roleService.IsLegacyAdmin(c.Request.Context(), adminIDStr)

	
	if !isLegacyAdmin {
		adminRoles, _ := h.roleService.GetUserRoles(c.Request.Context(), adminIDStr)
		targetRoles, _ := h.roleService.GetUserRoles(c.Request.Context(), userID)
		
		adminHighest := models.GetHighestRole(adminRoles)
		targetHighest := models.GetHighestRole(targetRoles)
		
		adminLevel := 0
		targetLevel := 0
		if adminHighest != nil {
			adminLevel = models.RoleHierarchy[*adminHighest]
		}
		if targetHighest != nil {
			targetLevel = models.RoleHierarchy[*targetHighest]
		}
		
		if adminLevel <= targetLevel && adminIDStr != userID {
			errorMsg := fmt.Sprintf("Cannot modify roles of users with same or higher hierarchy (Legacy=%v, AdminID=%s, Lvl=%d vs %d)", isLegacyAdmin, adminIDStr, adminLevel, targetLevel)
			if legacyErr != nil {
				errorMsg += fmt.Sprintf(" Err=%v", legacyErr)
			}
			c.JSON(http.StatusForbidden, gin.H{
				"error": errorMsg,
			})
			return
		}
	}

	err := h.roleService.GrantRole(c.Request.Context(), userID, role, &adminIDStr, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add role", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role added successfully", "role": req.Role})
}

// RemoveUserRole removes a role from a user
// DELETE /admin/users/:id/roles/:role
func (h *RoleHandler) RemoveUserRole(c *gin.Context) {
	userID := c.Param("id")
	roleStr := c.Param("role")

	if userID == "" || roleStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID and role are required"})
		return
	}

	role := models.RoleType(roleStr)
	if !role.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role type"})
		return
	}

	// Get admin user ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminIDStr, ok := adminID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin ID"})
		return
	}

	// Hierarchy check: admin can only modify users with lower hierarchy (unless legacy admin)
	isLegacyAdmin, legacyErr := h.roleService.IsLegacyAdmin(c.Request.Context(), adminIDStr)

	
	if !isLegacyAdmin {
		adminRoles, _ := h.roleService.GetUserRoles(c.Request.Context(), adminIDStr)
		targetRoles, _ := h.roleService.GetUserRoles(c.Request.Context(), userID)
		
		adminHighest := models.GetHighestRole(adminRoles)
		targetHighest := models.GetHighestRole(targetRoles)
		
		adminLevel := 0
		targetLevel := 0
		if adminHighest != nil {
			adminLevel = models.RoleHierarchy[*adminHighest]
		}
		if targetHighest != nil {
			targetLevel = models.RoleHierarchy[*targetHighest]
		}
		
		if adminLevel <= targetLevel && adminIDStr != userID {
			errorMsg := fmt.Sprintf("Cannot modify roles of users with same or higher hierarchy (Legacy=%v, AdminID=%s, Lvl=%d vs %d)", isLegacyAdmin, adminIDStr, adminLevel, targetLevel)
			if legacyErr != nil {
				errorMsg += fmt.Sprintf(" Err=%v", legacyErr)
			}
			c.JSON(http.StatusForbidden, gin.H{
				"error": errorMsg,
			})
			return
		}
	}

	err := h.roleService.RevokeRole(c.Request.Context(), userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

// GetUserRoles gets all roles for a user
// GET /admin/users/:id/roles
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	roles, err := h.roleService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}
