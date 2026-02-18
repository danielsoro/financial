package database

import (
	"context"
	"sync"

	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantCache struct {
	mu       sync.RWMutex
	byID     map[uuid.UUID]*entity.Tenant
	byDomain map[string]*entity.Tenant
}

func NewTenantCache() *TenantCache {
	return &TenantCache{
		byID:     make(map[uuid.UUID]*entity.Tenant),
		byDomain: make(map[string]*entity.Tenant),
	}
}

func (tc *TenantCache) Load(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx,
		`SELECT id, name, domain, schema_name, is_active, created_at, updated_at FROM tenants WHERE is_active = true`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	byID := make(map[uuid.UUID]*entity.Tenant)
	byDomain := make(map[string]*entity.Tenant)

	for rows.Next() {
		var t entity.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Domain, &t.SchemaName, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return err
		}
		byID[t.ID] = &t
		byDomain[t.Domain] = &t
	}

	tc.mu.Lock()
	tc.byID = byID
	tc.byDomain = byDomain
	tc.mu.Unlock()

	return nil
}

func (tc *TenantCache) GetByID(id uuid.UUID) (*entity.Tenant, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	t, ok := tc.byID[id]
	return t, ok
}

func (tc *TenantCache) GetByDomain(domain string) (*entity.Tenant, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	t, ok := tc.byDomain[domain]
	return t, ok
}
