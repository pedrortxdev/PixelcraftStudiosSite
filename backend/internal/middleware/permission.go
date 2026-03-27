package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

// RequirePermission middleware que verifica se o usuário tem uma permissão específica
func RequirePermission(permService *service.PermissionService, resource models.ResourceType, action models.ActionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// Verificar permissão
		hasPermission, err := permService.CheckPermission(userIDStr, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "Forbidden",
				"message":  "You don't have permission to perform this action",
				"resource": resource,
				"action":   action,
			})
			c.Abort()
			return
		}

		// Adicionar permissões ao contexto para uso posterior
		perms, _ := permService.GetUserPermissions(userIDStr)
		c.Set("user_permissions", perms)

		c.Next()
	}
}

// RequireAnyPermission middleware que verifica se o usuário tem pelo menos uma das permissões
func RequireAnyPermission(permService *service.PermissionService, checks []models.PermissionCheck) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// Verificar se tem pelo menos uma permissão
		hasAnyPermission := false
		for _, check := range checks {
			hasPermission, err := permService.CheckPermission(userIDStr, check.Resource, check.Action)
			if err == nil && hasPermission {
				hasAnyPermission = true
				break
			}
		}

		if !hasAnyPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You don't have permission to perform this action",
			})
			c.Abort()
			return
		}

		// Adicionar permissões ao contexto
		perms, _ := permService.GetUserPermissions(userIDStr)
		c.Set("user_permissions", perms)

		c.Next()
	}
}

// LoadUserPermissions middleware que carrega as permissões do usuário no contexto
func LoadUserPermissions(permService *service.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			if userIDStr, ok := userID.(string); ok {
				perms, _ := permService.GetUserPermissions(userIDStr)
				c.Set("user_permissions", perms)
			}
		}
		c.Next()
	}
}
