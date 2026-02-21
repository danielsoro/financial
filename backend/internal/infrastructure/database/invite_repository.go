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

type InviteRepo struct {
	pool *pgxpool.Pool
}

func NewInviteRepo(pool *pgxpool.Pool) *InviteRepo {
	return &InviteRepo{pool: pool}
}

func (r *InviteRepo) Create(ctx context.Context, invite *entity.Invite) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO invites (tenant_id, email, role, token, invited_by, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (tenant_id, email) DO UPDATE SET
		   role = EXCLUDED.role,
		   token = EXCLUDED.token,
		   invited_by = EXCLUDED.invited_by,
		   expires_at = EXCLUDED.expires_at,
		   accepted_at = NULL,
		   created_at = NOW()
		 RETURNING id, created_at`,
		invite.TenantID, invite.Email, invite.Role, invite.Token, invite.InvitedBy, invite.ExpiresAt,
	).Scan(&invite.ID, &invite.CreatedAt)
	return err
}

func (r *InviteRepo) FindByToken(ctx context.Context, token string) (*entity.Invite, error) {
	var i entity.Invite
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, email, role, token, invited_by, accepted_at, expires_at, created_at
		 FROM invites WHERE token = $1`, token,
	).Scan(&i.ID, &i.TenantID, &i.Email, &i.Role, &i.Token, &i.InvitedBy, &i.AcceptedAt, &i.ExpiresAt, &i.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &i, nil
}

func (r *InviteRepo) FindByTenantAndEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.Invite, error) {
	var i entity.Invite
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, email, role, token, invited_by, accepted_at, expires_at, created_at
		 FROM invites WHERE tenant_id = $1 AND email = $2`, tenantID, email,
	).Scan(&i.ID, &i.TenantID, &i.Email, &i.Role, &i.Token, &i.InvitedBy, &i.AcceptedAt, &i.ExpiresAt, &i.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &i, nil
}

func (r *InviteRepo) MarkAccepted(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE invites SET accepted_at = NOW() WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *InviteRepo) FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Invite, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, email, role, token, invited_by, accepted_at, expires_at, created_at
		 FROM invites WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []entity.Invite
	for rows.Next() {
		var i entity.Invite
		if err := rows.Scan(&i.ID, &i.TenantID, &i.Email, &i.Role, &i.Token, &i.InvitedBy, &i.AcceptedAt, &i.ExpiresAt, &i.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, i)
	}
	if invites == nil {
		invites = []entity.Invite{}
	}
	return invites, nil
}
