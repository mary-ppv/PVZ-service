package auth

import "github.com/dgrijalva/jwt-go"

type UserClaims struct {
	Role string `json:"role"`
	jwt.StandardClaims
}
