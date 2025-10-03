package middleware

import (
	"PVZ/pkg/helper"
	"PVZ/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := helper.GetUserRole(c)
		if role == "" {
			logger.Log.Printf("Role not found in context")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		
		allowed := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			logger.Log.Printf("Access denied for role: %s, allowed: %v", role, allowedRoles)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":          "access denied for your role",
				"required_roles": allowedRoles,
				"your_role":      role,
			})
			return
		}

		c.Next()
	}
}
