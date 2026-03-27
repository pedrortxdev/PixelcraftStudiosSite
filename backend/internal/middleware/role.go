package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/models"
)

// RoleMiddleware creates a middleware that verifies if the user has one of the allowed roles
// This middleware MUST be used after AuthMiddleware which sets "user_id" in context
func RoleMiddleware(db *sql.DB, allowedRoles ...models.RoleType) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User not authenticated"})
			c.Abort()
			return
		}

		// Get user's roles from database
		roles, err := getUserRoles(db, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user roles"})
			c.Abort()
			return
		}

		// Check if user has any of the allowed roles
		hasPermission := false
		for _, userRole := range roles {
			for _, allowed := range allowedRoles {
				if userRole == allowed {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			// Check if user is legacy admin (is_admin = true)
			isLegacy, err := isLegacyAdmin(db, userID.(string))
			if err == nil && isLegacy {
				hasPermission = true
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions"})
			c.Abort()
			return
		}

		// Set roles in context for handlers
		c.Set("user_roles", roles)
		c.Set("highest_role", getHighestRole(roles))
		c.Set("is_staff", true) // For compatibility with existing code

		c.Next()
	}
}

// AdminPanelMiddleware checks if user can access admin panel (any admin role)
func AdminPanelMiddleware(db *sql.DB) gin.HandlerFunc {
	return RoleMiddleware(db,
		models.RoleSupport,
		models.RoleAdmin,
		models.RoleDevelopment,
		models.RoleEngineering,
		models.RoleDirection,
	)
}

// CatalogEditMiddleware checks if user can edit catalog (plans, products)
func CatalogEditMiddleware(db *sql.DB) gin.HandlerFunc {
	return RoleMiddleware(db,
		models.RoleDevelopment,
		models.RoleEngineering,
		models.RoleDirection,
	)
}

// EmailManagementMiddleware checks if user can manage emails
func EmailManagementMiddleware(db *sql.DB) gin.HandlerFunc {
	return RoleMiddleware(db,
		models.RoleEngineering,
		models.RoleDirection,
	)
}

// FullAccessMiddleware checks if user has full access (Direction only)
func FullAccessMiddleware(db *sql.DB) gin.HandlerFunc {
	return RoleMiddleware(db, models.RoleDirection)
}

// getUserRoles fetches all active roles for a user from the database
func getUserRoles(db *sql.DB, userID string) ([]models.RoleType, error) {
	query := `
		SELECT role FROM user_roles 
		WHERE user_id = $1 
		AND (expires_at IS NULL OR expires_at > NOW())
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.RoleType
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, models.RoleType(role))
	}
	return roles, nil
}

// isLegacyAdmin checks if user has is_admin=true flag
func isLegacyAdmin(db *sql.DB, userID string) (bool, error) {
	var isAdmin bool
	query := `SELECT COALESCE(is_admin, false) FROM users WHERE id = $1`
	err := db.QueryRow(query, userID).Scan(&isAdmin)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}



// getHighestRole returns the highest role from a list
func getHighestRole(roles []models.RoleType) *models.RoleType {
	return models.GetHighestRole(roles)
}

// GetUserRolesFromContext retrieves user roles from the gin context
func GetUserRolesFromContext(c *gin.Context) []models.RoleType {
	roles, exists := c.Get("user_roles")
	if !exists {
		return nil
	}
	return roles.([]models.RoleType)
}

// HasRoleInContext checks if the user in context has a specific role
func HasRoleInContext(c *gin.Context, role models.RoleType) bool {
	roles := GetUserRolesFromContext(c)
	return models.HasRole(roles, role)
}

// IsStaffInContext checks if the user in context is staff (has admin access)
func IsStaffInContext(c *gin.Context) bool {
	_, exists := c.Get("is_staff")
	return exists
}

// CanEditUserMiddleware checks if the current user can edit the target user
// This is used for password/balance editing where Engineering can only edit lower roles
func CanEditUserMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		targetUserID := c.Param("id")

		if targetUserID == "" {
			c.Next()
			return
		}

		sourceRoles := GetUserRolesFromContext(c)
		
		// Direction can edit anyone
		if models.HasRole(sourceRoles, models.RoleDirection) {
			c.Next()
			return
		}

		// Get target user's roles
		targetRoles, err := getUserRoles(db, targetUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check target user roles"})
			c.Abort()
			return
		}

		// Check if source can edit target
		if !models.CanEditRole(sourceRoles, targetRoles) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Cannot edit user with equal or higher role"})
			c.Abort()
			return
		}

		// Cannot edit self
		if userIDStr, ok := userID.(string); ok && userIDStr == targetUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Cannot edit your own account via admin"})
			c.Abort()
			return
		}

		c.Next()
	}
}
