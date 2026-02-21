package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SchemaManager struct {
	pool *pgxpool.Pool
}

func NewSchemaManager(pool *pgxpool.Pool) *SchemaManager {
	return &SchemaManager{pool: pool}
}

// InitAllTenants initializes schemas for all existing active tenants in the DB.
func (sm *SchemaManager) InitAllTenants(ctx context.Context, databaseURL, migrationsDir string) error {
	rows, err := sm.pool.Query(ctx,
		`SELECT schema_name FROM tenants WHERE is_active = true`,
	)
	if err != nil {
		return fmt.Errorf("querying tenants: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return err
		}
		schemas = append(schemas, schemaName)
	}

	for _, schemaName := range schemas {
		if err := sm.InitTenantSchema(ctx, databaseURL, migrationsDir, schemaName); err != nil {
			return fmt.Errorf("initializing tenant %s: %w", schemaName, err)
		}
		log.Printf("Tenant schema %s: ready", schemaName)
	}

	return nil
}

// InitTenantSchema creates a schema (if not exists) and runs tenant migrations.
// Public so it can be called by the registration flow.
func (sm *SchemaManager) InitTenantSchema(ctx context.Context, databaseURL, migrationsDir, schemaName string) error {
	if !validSchemaName.MatchString(schemaName) {
		return fmt.Errorf("invalid schema name: %s", schemaName)
	}

	_, err := sm.pool.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+pgx.Identifier{schemaName}.Sanitize())
	if err != nil {
		return fmt.Errorf("creating schema: %w", err)
	}

	if err := sm.runTenantMigrations(databaseURL, migrationsDir, schemaName); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

func (sm *SchemaManager) runTenantMigrations(databaseURL, migrationsDir, schemaName string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("parsing database URL: %w", err)
	}
	q := u.Query()
	q.Set("search_path", schemaName)
	u.RawQuery = q.Encode()

	m, err := migrate.New("file://"+migrationsDir, u.String())
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}
