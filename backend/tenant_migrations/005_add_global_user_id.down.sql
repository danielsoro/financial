DROP INDEX IF EXISTS idx_users_global_user_id;
ALTER TABLE users DROP COLUMN IF EXISTS global_user_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'user'));
