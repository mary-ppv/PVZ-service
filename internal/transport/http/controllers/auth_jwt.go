package controllers

import (
	"PVZ/pkg/auth"
	"PVZ/pkg/logger"
	"errors"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func ParseJWT(tokenString string, jwtKey []byte) (string, error) {
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	token, err := jwt.ParseWithClaims(tokenString, &auth.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*auth.UserClaims); ok && token.Valid {
		logger.Log.Printf("ParseJWT: successfully extracted role=%s", claims.Role)
		return claims.Role, nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		role, ok := claims["role"].(string)
		if !ok {
			return "", errors.New("role not found in token")
		}
		logger.Log.Printf("ParseJWT: extracted role=%s from MapClaims", role)
		return role, nil
	}

	return "", errors.New("invalid token claims")
}
