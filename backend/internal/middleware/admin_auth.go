package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware checks if the authenticated user is an admin
// It MUST be used after AuthMiddleware which sets "user_id" in context
func AdminAuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User not authenticated"})
			c.Abort()
			return
		}

		// Check if user is admin in database
		var isAdmin bool
		err := db.QueryRow("SELECT is_admin FROM users WHERE id = $1", userID).Scan(&isAdmin)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking admin status"})
			}
			c.Abort()
			return
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access required"})
			c.Abort()
			return
		}

		// Set is_admin in context for handlers that need it
		c.Set("is_admin", true)

		c.Next()
	}
}
