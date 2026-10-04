-- +goose Up

-- 1. Enhance Accounts table
ALTER TABLE accounts ADD COLUMN nickname TEXT;
ALTER TABLE accounts ADD COLUMN customer_id TEXT;
ALTER TABLE accounts ADD COLUMN ifsc_code TEXT;
ALTER TABLE accounts ADD COLUMN branch_name TEXT;

-- 2. Enhance Transactions table for Indian ecosystem & clean analytics
ALTER TABLE transactions ADD COLUMN upi_vpa TEXT;
ALTER TABLE transactions ADD COLUMN card_last4 TEXT;
ALTER TABLE transactions ADD COLUMN is_transfer BOOLEAN DEFAULT 0;
ALTER TABLE transactions ADD COLUMN is_excluded BOOLEAN DEFAULT 0;
ALTER TABLE transactions ADD COLUMN original_currency TEXT;
ALTER TABLE transactions ADD COLUMN original_amount REAL;

-- 3. Enhance Statement Imports log with financial totals
ALTER TABLE statement_imports ADD COLUMN opening_balance REAL;
ALTER TABLE statement_imports ADD COLUMN closing_balance REAL;
ALTER TABLE statement_imports ADD COLUMN total_debits REAL;
ALTER TABLE statement_imports ADD COLUMN total_credits REAL;

-- 4. Credit Card Monthly Billing Statements Table
CREATE TABLE IF NOT EXISTS credit_card_bills (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    statement_import_id TEXT,
    statement_date DATE NOT NULL,
    payment_due_date DATE NOT NULL,
    total_due_amount REAL NOT NULL,
    minimum_due_amount REAL,
    reward_points_earned REAL DEFAULT 0,
    reward_points_balance REAL DEFAULT 0,
    payment_status TEXT DEFAULT 'UNPAID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(account_id) REFERENCES accounts(id),
    FOREIGN KEY(statement_import_id) REFERENCES statement_imports(id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_is_transfer ON transactions(is_transfer);
CREATE INDEX IF NOT EXISTS idx_transactions_upi_vpa ON transactions(upi_vpa);
CREATE INDEX IF NOT EXISTS idx_cc_bills_account_id ON credit_card_bills(account_id);

-- +goose Down
DROP INDEX IF EXISTS idx_cc_bills_account_id;
DROP INDEX IF EXISTS idx_transactions_upi_vpa;
DROP INDEX IF EXISTS idx_transactions_is_transfer;
DROP TABLE IF EXISTS credit_card_bills;
ALTER TABLE statement_imports DROP COLUMN total_credits;
ALTER TABLE statement_imports DROP COLUMN total_debits;
ALTER TABLE statement_imports DROP COLUMN closing_balance;
ALTER TABLE statement_imports DROP COLUMN opening_balance;
ALTER TABLE transactions DROP COLUMN original_amount;
ALTER TABLE transactions DROP COLUMN original_currency;
ALTER TABLE transactions DROP COLUMN is_excluded;
ALTER TABLE transactions DROP COLUMN is_transfer;
ALTER TABLE transactions DROP COLUMN card_last4;
ALTER TABLE transactions DROP COLUMN upi_vpa;
ALTER TABLE accounts DROP COLUMN branch_name;
ALTER TABLE accounts DROP COLUMN ifsc_code;
ALTER TABLE accounts DROP COLUMN customer_id;
ALTER TABLE accounts DROP COLUMN nickname;
