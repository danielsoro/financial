DELETE FROM categories WHERE is_default = true AND user_id IS NULL;
