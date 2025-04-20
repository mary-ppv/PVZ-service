package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type UserClaims struct {
	Role string `json:"role"`
	jwt.StandardClaims
}

// AuthMiddleware validates JWT token and user role
func AuthMiddleware(jwtKey []byte, logger *log.Logger, allowedRoles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Println("Received request to authenticate")

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Println("Missing authorization header")
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				logger.Println("Invalid authorization header format")
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})
			if err != nil || !token.Valid {
				logger.Printf("Invalid token: %v", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(*jwt.MapClaims)
			if !ok {
				logger.Println("Invalid token claims")
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			role, ok := (*claims)["role"].(string)
			if !ok {
				logger.Println("Missing or invalid role in token")
				http.Error(w, "Missing or invalid role in token", http.StatusUnauthorized)
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
				logger.Printf("Access denied for role: %s", role)
				http.Error(w, "Access denied for this role", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), "userRole", role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
