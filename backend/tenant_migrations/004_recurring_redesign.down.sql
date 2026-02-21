-- Recreate recurring_occurrences table
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

-- Remove paused_at from recurring_transactions
ALTER TABLE recurring_transactions DROP COLUMN IF EXISTS paused_at;

-- Remove recurring_id from transactions
DROP INDEX IF EXISTS idx_transactions_recurring_id;
ALTER TABLE transactions DROP COLUMN IF EXISTS recurring_id;
