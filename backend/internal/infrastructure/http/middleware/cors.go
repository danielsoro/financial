package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin != "" && isAllowedOrigin(origin, allowedOrigin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isAllowedOrigin(origin, allowedOrigin string) bool {
	if allowedOrigin == "" || allowedOrigin == "*" {
		return true
	}

	// allowedOrigin = "dnafami.com.br"
	// Allow: https://dnafami.com.br, https://*.dnafami.com.br
	origin = strings.TrimPrefix(origin, "https://")
	origin = strings.TrimPrefix(origin, "http://")

	if origin == allowedOrigin {
		return true
	}

	if strings.HasSuffix(origin, "."+allowedOrigin) {
		return true
	}

	// Allow localhost for development
	if strings.HasPrefix(origin, "localhost") {
		return true
	}

	return false
}
