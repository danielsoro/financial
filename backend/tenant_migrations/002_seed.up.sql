INSERT INTO categories (name, type, is_default)
SELECT name, type, true
FROM (VALUES
    ('Alimentação', 'expense'),
    ('Transporte', 'expense'),
    ('Moradia', 'expense'),
    ('Saúde', 'expense'),
    ('Educação', 'expense'),
    ('Lazer', 'expense'),
    ('Salário', 'income'),
    ('Freelance', 'income'),
    ('Investimentos', 'both'),
    ('Outros', 'both')
) AS v(name, type)
WHERE NOT EXISTS (
    SELECT 1 FROM categories LIMIT 1
);
