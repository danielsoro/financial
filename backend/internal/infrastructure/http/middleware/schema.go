package middleware

import (
	"net/http"

	"github.com/dcunha/finance/backend/internal/infrastructure/database"
	"github.com/dcunha/finance/backend/internal/tenant"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SchemaConn(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		schema := tenant.SchemaFromContext(c.Request.Context())
		if schema == "" {
			c.Next()
			return
		}

		conn, release, err := database.AcquireWithSchema(c.Request.Context(), pool)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		defer release()

		ctx := database.ContextWithConn(c.Request.Context(), conn)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
