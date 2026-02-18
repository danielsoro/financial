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
