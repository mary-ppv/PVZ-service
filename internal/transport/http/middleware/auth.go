package middleware

import (
	"PVZ/internal/transport/http/controllers"
	"github.com/gin-gonic/gin"
	"log"
)

func JWTMiddleware(jwtKey []byte, logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString != "" {
			role, err := controllers.ParseJWT(tokenString, jwtKey)
			if err != nil {
				logger.Printf("Invalid token: %v", err)
				c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
				return
			}
			c.Set("userRole", role)
		}
		c.Next()
	}
}
