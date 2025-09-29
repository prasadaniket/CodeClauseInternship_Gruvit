package middleware

import (
	"net/http"
	"strings"

	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates middleware that validates tokens with the Java auth service
func AuthMiddleware(authClient *services.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Validate token with Java auth service
		authResp, err := authClient.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token validation failed",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !authResp.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"details": authResp.Error,
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", authResp.UserID)
		c.Set("username", authResp.Username)
		c.Set("role", authResp.Role)
		c.Set("token", tokenString)

		c.Next()
	}
}

// OptionalAuthMiddleware creates middleware that validates tokens if present
func OptionalAuthMiddleware(authClient *services.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Remove "Bearer " prefix if present
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		// Validate token with Java auth service
		authResp, err := authClient.ValidateToken(tokenString)
		if err != nil {
			// Log error but don't abort for optional auth
			c.Set("auth_error", err.Error())
			c.Next()
			return
		}

		if authResp.Valid {
			// Set user information in context
			c.Set("user_id", authResp.UserID)
			c.Set("username", authResp.Username)
			c.Set("role", authResp.Role)
			c.Set("token", tokenString)
			c.Set("authenticated", true)
		}

		c.Next()
	}
}

// AdminMiddleware creates middleware that requires admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserMiddleware creates middleware that requires user role or higher
func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "USER" && role != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "User access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
