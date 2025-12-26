package main

import (
	"strings"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequirePermission(requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization") // [cite: 23, 72]
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return JWTSecret, nil 
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			perms, exists := claims["permissions"].([]interface{})
			if !exists {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No permissions found in token"})
				return
			}

			hasPerm := false
			for _, p := range perms {
				if p.(string) == requiredPerm {
					hasPerm = true
					break
				}
			}

			if !hasPerm {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Missing required permission: " + requiredPerm})
				return
			}
		}

		c.Next()
	}
}