package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quochao170402/ecommerce-aws/user-service/auth"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("user_id", claims[auth.ClaimUserID])
		c.Set("email", claims[auth.ClaimUserEmail])
		c.Set("name", claims[auth.ClaimUserName])
		c.Set("role", claims[auth.ClaimRole])

		c.Next()
	}
}

func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		role, ok := roleValue.(string)
		if !ok || role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden – requires " + requiredRole})
			return
		}

		c.Next()
	}
}
