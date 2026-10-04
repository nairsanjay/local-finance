-- +goose Up

-- 1. Enhance Transactions with reconciliation peer links, match reasons, and net offset amounts
ALTER TABLE transactions ADD COLUMN transfer_peer_id TEXT;
ALTER TABLE transactions ADD COLUMN transfer_match_reason TEXT;
ALTER TABLE transactions ADD COLUMN net_amount REAL;

-- 2. Indexes for fast transfer pairing and merchant deep-dive aggregation
CREATE INDEX IF NOT EXISTS idx_transactions_transfer_peer ON transactions(transfer_peer_id);
CREATE INDEX IF NOT EXISTS idx_transactions_cleaned_payee ON transactions(cleaned_payee);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_cleaned_payee;
DROP INDEX IF EXISTS idx_transactions_transfer_peer;
ALTER TABLE transactions DROP COLUMN net_amount;
ALTER TABLE transactions DROP COLUMN transfer_match_reason;
ALTER TABLE transactions DROP COLUMN transfer_peer_id;
