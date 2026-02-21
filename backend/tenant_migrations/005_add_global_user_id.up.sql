ALTER TABLE users ADD COLUMN IF NOT EXISTS global_user_id UUID;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('owner', 'admin', 'user'));
CREATE INDEX IF NOT EXISTS idx_users_global_user_id ON users(global_user_id);
