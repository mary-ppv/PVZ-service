package middleware

import (
	"PVZ/internal/transport/http/controllers"
	"PVZ/pkg/helper"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware(jwtKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		role, err := controllers.ParseJWT(tokenString, jwtKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		helper.SetUserRole(c, role)

		c.Next()
	}
}
