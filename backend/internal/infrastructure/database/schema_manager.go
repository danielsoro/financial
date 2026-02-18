package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type SchemaManager struct {
	pool *pgxpool.Pool
}

func NewSchemaManager(pool *pgxpool.Pool) *SchemaManager {
	return &SchemaManager{pool: pool}
}

func (sm *SchemaManager) InitAllTenants(ctx context.Context, databaseURL, migrationsDir string) error {
	rows, err := sm.pool.Query(ctx,
		`SELECT domain, schema_name FROM tenants WHERE is_active = true`,
	)
	if err != nil {
		return fmt.Errorf("querying tenants: %w", err)
	}
	defer rows.Close()

	type tenantInfo struct {
		domain     string
		schemaName string
	}
	var tenants []tenantInfo
	for rows.Next() {
		var t tenantInfo
		if err := rows.Scan(&t.domain, &t.schemaName); err != nil {
			return err
		}
		tenants = append(tenants, t)
	}

	for _, t := range tenants {
		if err := sm.initTenant(ctx, databaseURL, migrationsDir, t.schemaName); err != nil {
			return fmt.Errorf("initializing tenant %s: %w", t.domain, err)
		}
		log.Printf("Tenant schema %s: ready", t.schemaName)
	}

	return nil
}

func (sm *SchemaManager) initTenant(ctx context.Context, databaseURL, migrationsDir, schemaName string) error {
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

	if err := sm.seedAdminIfEmpty(ctx, schemaName); err != nil {
		return fmt.Errorf("seeding admin: %w", err)
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

func (sm *SchemaManager) seedAdminIfEmpty(ctx context.Context, schemaName string) error {
	var count int
	err := sm.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s.users", pgx.Identifier{schemaName}.Sanitize()),
	).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = sm.pool.Exec(ctx,
		fmt.Sprintf(
			"INSERT INTO %s.users (name, email, password_hash, role) VALUES ($1, $2, $3, $4)",
			pgx.Identifier{schemaName}.Sanitize(),
		),
		"Admin", "admin@admin.com", string(hash), "admin",
	)
	if err != nil {
		return err
	}

	log.Printf("Tenant schema %s: seeded admin user", schemaName)
	return nil
}

func EnsureTenantsFromEnv(ctx context.Context, pool *pgxpool.Pool, tenantList string) error {
	if tenantList == "" {
		return nil
	}

	for _, domain := range strings.Split(tenantList, ",") {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}

		schemaName := "tenant_" + strings.ReplaceAll(domain, "-", "_")
		name := strings.ToUpper(domain[:1]) + domain[1:]

		_, err := pool.Exec(ctx,
			`INSERT INTO tenants (name, domain, schema_name) VALUES ($1, $2, $3) ON CONFLICT (domain) DO NOTHING`,
			name, domain, schemaName,
		)
		if err != nil {
			return fmt.Errorf("ensuring tenant %s: %w", domain, err)
		}
	}

	return nil
}
