-- +goose Up
ALTER TABLE transactions ADD COLUMN is_manual_category BOOLEAN DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_transactions_manual_category ON transactions(is_manual_category);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_manual_category;
ALTER TABLE transactions DROP COLUMN is_manual_category;
