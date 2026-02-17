DROP INDEX IF EXISTS idx_categories_parent_id;
DROP INDEX IF EXISTS idx_categories_unique_child;
DROP INDEX IF EXISTS idx_categories_unique_root;

-- Restore original unique constraint
ALTER TABLE categories ADD CONSTRAINT categories_user_id_name_key UNIQUE(user_id, name);

ALTER TABLE categories DROP COLUMN IF EXISTS parent_id;
