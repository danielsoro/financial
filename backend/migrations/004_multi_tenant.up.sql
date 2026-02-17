-- 1. Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(100) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Seed default tenant
INSERT INTO tenants (name, domain) VALUES ('Financial', 'financial');

-- 3. Alter users: add tenant_id and role
ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id);
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user'
    CHECK (role IN ('admin', 'user'));

-- 4. Backfill existing users to default tenant
UPDATE users SET tenant_id = (SELECT id FROM tenants WHERE domain = 'financial');

-- 5. Seed super_admin user (tenant_id = NULL — not bound to any tenant)
INSERT INTO users (name, email, password_hash, tenant_id, role)
VALUES (
    'Admin',
    'admin@admin.com',
    '$2a$10$ReG8EsupKR.29AVE/7O02OIQpG9y1b2CygBtg4FB8T/nlpx3ex/kG',
    (SELECT id FROM tenants WHERE domain = 'financial'),
    'admin'
);

-- 6. Add tenant_id to categories, transactions, expense_limits
ALTER TABLE categories ADD COLUMN tenant_id UUID REFERENCES tenants(id);
ALTER TABLE transactions ADD COLUMN tenant_id UUID REFERENCES tenants(id);
ALTER TABLE expense_limits ADD COLUMN tenant_id UUID REFERENCES tenants(id);

-- 7. Backfill tenant_id from user's tenant
UPDATE categories SET tenant_id = (
    SELECT u.tenant_id FROM users u WHERE u.id = categories.user_id
) WHERE user_id IS NOT NULL;

-- Default categories (user_id IS NULL) go to default tenant
UPDATE categories SET tenant_id = (SELECT id FROM tenants WHERE domain = 'financial')
WHERE user_id IS NULL;

UPDATE transactions SET tenant_id = (
    SELECT u.tenant_id FROM users u WHERE u.id = transactions.user_id
);

UPDATE expense_limits SET tenant_id = (
    SELECT u.tenant_id FROM users u WHERE u.id = expense_limits.user_id
);

-- 8. Make tenant_id NOT NULL (except users — super_admin has NULL)
ALTER TABLE categories ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE transactions ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE expense_limits ALTER COLUMN tenant_id SET NOT NULL;

-- 9. Recreate unique constraints with tenant scope

-- Categories: drop old unique indices and create tenant-scoped ones
DROP INDEX IF EXISTS idx_categories_unique_root;
DROP INDEX IF EXISTS idx_categories_unique_child;

CREATE UNIQUE INDEX idx_categories_unique_root
    ON categories (tenant_id, name) WHERE parent_id IS NULL;

CREATE UNIQUE INDEX idx_categories_unique_child
    ON categories (tenant_id, parent_id, name) WHERE parent_id IS NOT NULL;

-- Expense limits: drop old constraints and recreate with tenant scope
ALTER TABLE expense_limits DROP CONSTRAINT IF EXISTS expense_limits_user_id_category_id_month_year_key;
DROP INDEX IF EXISTS idx_expense_limits_global;

CREATE UNIQUE INDEX idx_expense_limits_tenant_cat
    ON expense_limits (tenant_id, category_id, month, year);

CREATE UNIQUE INDEX idx_expense_limits_tenant_global
    ON expense_limits (tenant_id, month, year)
    WHERE category_id IS NULL;

-- 10. Create indices
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_categories_tenant_id ON categories(tenant_id);
CREATE INDEX idx_transactions_tenant_id ON transactions(tenant_id);
CREATE INDEX idx_expense_limits_tenant_id ON expense_limits(tenant_id);
