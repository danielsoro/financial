-- Remove super_admin role: all users must now belong to a tenant

-- 1. Delete any super_admin users (safety net — CHECK constraint already prevents creation)
DELETE FROM users WHERE role = 'super_admin';

-- 2. Backfill any users with NULL tenant_id to the default tenant
UPDATE users SET tenant_id = (SELECT id FROM tenants WHERE domain = 'financial')
WHERE tenant_id IS NULL;

-- 3. Make tenant_id NOT NULL
ALTER TABLE users ALTER COLUMN tenant_id SET NOT NULL;
