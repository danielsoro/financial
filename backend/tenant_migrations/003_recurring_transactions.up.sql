-- Recurring Transactions (templates)
CREATE TABLE IF NOT EXISTS recurring_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id),
    type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    amount DECIMAL(12,2) NOT NULL CHECK (amount > 0),
    description VARCHAR(500),
    frequency VARCHAR(10) NOT NULL CHECK (frequency IN ('weekly', 'biweekly', 'monthly', 'yearly')),
    start_date DATE NOT NULL,
    end_date DATE,
    max_occurrences INT,
    day_of_month INT CHECK (day_of_month BETWEEN 1 AND 31),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recurring_transactions_user_id ON recurring_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_recurring_transactions_active ON recurring_transactions(user_id) WHERE is_active = true;

-- Recurring Occurrences (generated instances)
CREATE TABLE IF NOT EXISTS recurring_occurrences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recurring_id UUID NOT NULL REFERENCES recurring_transactions(id) ON DELETE CASCADE,
    occurrence_date DATE NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'skipped')),
    amount DECIMAL(12,2) NOT NULL CHECK (amount > 0),
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_recurring_occurrences_unique ON recurring_occurrences(recurring_id, occurrence_date);
CREATE INDEX IF NOT EXISTS idx_recurring_occurrences_recurring_id ON recurring_occurrences(recurring_id);
CREATE INDEX IF NOT EXISTS idx_recurring_occurrences_pending ON recurring_occurrences(recurring_id) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_recurring_occurrences_date ON recurring_occurrences(occurrence_date);
