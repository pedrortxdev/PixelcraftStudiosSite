package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Auth Middleware: Verificando requisição...")

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("Auth Error: Header Authorization não encontrado")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization header",
			})
			c.Abort()
			return
		}

		log.Printf("Auth Middleware: Authorization header received")

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("Auth Error: Formato Bearer inválido")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("Auth Error: Invalid signing method")
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			log.Printf("Auth Error: Token expirado ou assinatura inválida - Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extract user ID from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("Auth Error: Invalid token claims format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			c.Abort()
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			log.Printf("Auth Error: Missing user_id in token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing user_id in token",
			})
			c.Abort()
			return
		}

		// Validate UUID format
		_, err = uuid.Parse(userIDStr)
		if err != nil {
			log.Printf("Auth Error: Invalid user_id format - Value: %s, Error: %v", userIDStr, err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user_id format",
			})
			c.Abort()
			return
		}

		// Store user ID as string in context for use in handlers
		c.Set("user_id", userIDStr)
		
		// Extract is_admin from claims if present
		if isAdmin, ok := claims["is_admin"].(bool); ok {
			c.Set("is_admin", isAdmin)
			log.Printf("Auth Success: Usuário autenticado ID: %s, Admin: %v", userIDStr, isAdmin)
		} else {
			log.Printf("Auth Success: Usuário autenticado ID: %s", userIDStr)
		}

		c.Next()
	}
}