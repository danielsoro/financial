package database

import (
	"context"
	"fmt"
	"regexp"

	"github.com/dcunha/finance/backend/internal/tenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var validSchemaName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type connKey struct{}

// ContextWithConn stores a schema-scoped database connection in the context.
// Used by the SchemaConn middleware and the login handler.
func ContextWithConn(ctx context.Context, conn *pgxpool.Conn) context.Context {
	return context.WithValue(ctx, connKey{}, conn)
}

// ConnFromContext retrieves the schema-scoped database connection from the context.
// Returns an error if no connection is found, indicating the SchemaConn middleware is missing.
func ConnFromContext(ctx context.Context) (*pgxpool.Conn, error) {
	conn, ok := ctx.Value(connKey{}).(*pgxpool.Conn)
	if !ok || conn == nil {
		return nil, fmt.Errorf("no database connection in context: ensure SchemaConn middleware is applied")
	}
	return conn, nil
}

func AcquireWithSchema(ctx context.Context, pool *pgxpool.Pool) (*pgxpool.Conn, func(), error) {
	schema := tenant.SchemaFromContext(ctx)
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("acquiring connection: %w", err)
	}

	if schema != "" {
		if !validSchemaName.MatchString(schema) {
			conn.Release()
			return nil, nil, fmt.Errorf("invalid schema name: %s", schema)
		}
		_, err = conn.Exec(ctx, "SET search_path TO "+pgx.Identifier{schema}.Sanitize())
		if err != nil {
			conn.Release()
			return nil, nil, fmt.Errorf("setting search_path: %w", err)
		}
	}

	return conn, func() {
		conn.Exec(context.Background(), "RESET search_path")
		conn.Release()
	}, nil
}
