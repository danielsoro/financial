-- Add recurring_id to transactions for reverse link
ALTER TABLE transactions ADD COLUMN recurring_id UUID REFERENCES recurring_transactions(id) ON DELETE SET NULL;
CREATE INDEX idx_transactions_recurring_id ON transactions(recurring_id);

-- Add paused_at to recurring_transactions
ALTER TABLE recurring_transactions ADD COLUMN paused_at TIMESTAMPTZ;

-- Drop recurring_occurrences table (no longer needed)
DROP TABLE IF EXISTS recurring_occurrences;
