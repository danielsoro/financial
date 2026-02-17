-- Drop indices
DROP INDEX IF EXISTS idx_expense_limits_tenant_id;
DROP INDEX IF EXISTS idx_transactions_tenant_id;
DROP INDEX IF EXISTS idx_categories_tenant_id;
DROP INDEX IF EXISTS idx_users_tenant_id;

-- Restore expense_limits constraints
DROP INDEX IF EXISTS idx_expense_limits_tenant_global;
DROP INDEX IF EXISTS idx_expense_limits_tenant_cat;

ALTER TABLE expense_limits ADD CONSTRAINT expense_limits_user_id_category_id_month_year_key
    UNIQUE(user_id, category_id, month, year);

CREATE UNIQUE INDEX idx_expense_limits_global
    ON expense_limits(user_id, month, year)
    WHERE category_id IS NULL;

-- Restore category constraints
DROP INDEX IF EXISTS idx_categories_unique_root;
DROP INDEX IF EXISTS idx_categories_unique_child;

CREATE UNIQUE INDEX idx_categories_unique_root
    ON categories (user_id, name) WHERE parent_id IS NULL;

CREATE UNIQUE INDEX idx_categories_unique_child
    ON categories (user_id, parent_id, name) WHERE parent_id IS NOT NULL;

-- Drop tenant_id columns
ALTER TABLE expense_limits DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE transactions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE categories DROP COLUMN IF EXISTS tenant_id;

-- Delete super_admin user
DELETE FROM users WHERE email = 'super@admin.com';

-- Drop role and tenant_id from users
ALTER TABLE users DROP COLUMN IF EXISTS role;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;

-- Drop tenants table
DROP TABLE IF EXISTS tenants;
