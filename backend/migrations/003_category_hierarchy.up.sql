ALTER TABLE categories ADD COLUMN parent_id UUID REFERENCES categories(id) ON DELETE CASCADE;

-- Replace UNIQUE(user_id, name) with constraints that allow same name under different parents
ALTER TABLE categories DROP CONSTRAINT categories_user_id_name_key;

CREATE UNIQUE INDEX idx_categories_unique_root
    ON categories (user_id, name) WHERE parent_id IS NULL;

CREATE UNIQUE INDEX idx_categories_unique_child
    ON categories (user_id, parent_id, name) WHERE parent_id IS NOT NULL;

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
