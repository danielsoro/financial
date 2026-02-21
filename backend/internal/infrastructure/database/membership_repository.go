package database

import (
	"context"
	"errors"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MembershipRepo struct {
	pool *pgxpool.Pool
}

func NewMembershipRepo(pool *pgxpool.Pool) *MembershipRepo {
	return &MembershipRepo{pool: pool}
}

func (r *MembershipRepo) Create(ctx context.Context, m *entity.Membership) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO memberships (global_user_id, tenant_id, schema_user_id, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		m.GlobalUserID, m.TenantID, m.SchemaUserID, m.Role,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrAlreadyMember
		}
		return err
	}
	return nil
}

func (r *MembershipRepo) FindByGlobalUser(ctx context.Context, globalUserID uuid.UUID) ([]entity.TenantMembership, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT m.tenant_id, t.name, m.role
		 FROM memberships m
		 JOIN tenants t ON t.id = m.tenant_id
		 WHERE m.global_user_id = $1 AND t.is_active = true
		 ORDER BY t.name ASC`, globalUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []entity.TenantMembership
	for rows.Next() {
		var tm entity.TenantMembership
		if err := rows.Scan(&tm.TenantID, &tm.TenantName, &tm.Role); err != nil {
			return nil, err
		}
		memberships = append(memberships, tm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if memberships == nil {
		memberships = []entity.TenantMembership{}
	}
	return memberships, nil
}

func (r *MembershipRepo) FindByGlobalUserAndTenant(ctx context.Context, globalUserID, tenantID uuid.UUID) (*entity.Membership, error) {
	var m entity.Membership
	err := r.pool.QueryRow(ctx,
		`SELECT id, global_user_id, tenant_id, schema_user_id, role, created_at, updated_at
		 FROM memberships WHERE global_user_id = $1 AND tenant_id = $2`, globalUserID, tenantID,
	).Scan(&m.ID, &m.GlobalUserID, &m.TenantID, &m.SchemaUserID, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *MembershipRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM memberships WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
