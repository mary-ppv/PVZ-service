package helper

import "github.com/gin-gonic/gin"

const userRoleKey = "userRole"

func SetUserRole(c *gin.Context, role string) {
	c.Set(userRoleKey, role)
}

func GetUserRole(c *gin.Context) string {
	if val, exists := c.Get(userRoleKey); exists {
		if role, ok := val.(string); ok {
			return role
		}
	}
	return ""
}
