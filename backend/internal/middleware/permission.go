package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

// RequirePermission middleware that checks if user has a specific permission
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

		// Parse UUID
		userIDUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		// Check permission (WITH CONTEXT + UUID)
		hasPermission, err := permService.HasPermission(c.Request.Context(), userIDUUID, resource, action)
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

		// Add permissions to context for later use
		perms, _ := permService.GetUserPermissions(c.Request.Context(), userIDUUID)
		c.Set("user_permissions", perms)

		c.Next()
	}
}

// RequireRole middleware that checks if user has a specific role
func RequireRole(permService *service.PermissionService, requiredRole string) gin.HandlerFunc {
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

		// Parse UUID
		userIDUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		// Get user permissions (which includes roles)
		perms, err := permService.GetUserPermissions(c.Request.Context(), userIDUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, role := range perms.Roles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":       "Forbidden",
				"message":     "You don't have the required role",
				"required":    requiredRole,
				"your_roles":  perms.Roles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware that checks if user has ANY of the specified roles
func RequireAnyRole(permService *service.PermissionService, roles ...string) gin.HandlerFunc {
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

		// Parse UUID
		userIDUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		// Get user roles
		perms, err := permService.GetUserPermissions(c.Request.Context(), userIDUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, userRole := range perms.Roles {
			for _, requiredRole := range roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":       "Forbidden",
				"message":     "You don't have any of the required roles",
				"required":    roles,
				"your_roles":  perms.Roles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
