package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireCorePermission(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms := c.GetHeader("X-Permissions")
		if perms == "" || !strings.Contains(perms, required) {
			c.AbortWithStatusJSON(403, gin.H{"error": "Missing permission: " + required})
			return
		}
		c.Next()
	}
}
