package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		if role != "admin" && role != "owner" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

func RequireOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		if role != "owner" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "owner access required"})
			return
		}
		c.Next()
	}
}
