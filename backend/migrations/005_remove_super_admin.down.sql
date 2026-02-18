-- Allow NULL tenant_id again
ALTER TABLE users ALTER COLUMN tenant_id DROP NOT NULL;
