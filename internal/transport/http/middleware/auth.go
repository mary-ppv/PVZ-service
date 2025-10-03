package middleware

import (
	"PVZ/internal/transport/http/controllers"
	"PVZ/pkg/helper"
	"PVZ/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware(jwtKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			logger.Log.Printf("Missing token in request")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		role, err := controllers.ParseJWT(tokenString, jwtKey)
		if err != nil {
			logger.Log.Printf("Invalid token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		helper.SetUserRole(c, role)

		c.Next()
	}
}
