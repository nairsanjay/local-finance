-- +goose Up

-- 1. Enhance Accounts table with credit card specific parameters
ALTER TABLE accounts ADD COLUMN annual_fee REAL DEFAULT 0.0;
ALTER TABLE accounts ADD COLUMN fee_waiver_threshold REAL DEFAULT 0.0;
ALTER TABLE accounts ADD COLUMN billing_day INTEGER DEFAULT 1;
ALTER TABLE accounts ADD COLUMN payment_due_days INTEGER DEFAULT 20;
ALTER TABLE accounts ADD COLUMN card_color TEXT DEFAULT '#1E293B';
ALTER TABLE accounts ADD COLUMN reward_type TEXT DEFAULT 'CASHBACK';
ALTER TABLE accounts ADD COLUMN base_reward_rate REAL DEFAULT 1.0;

-- 2. Card Reward Rules table
CREATE TABLE IF NOT EXISTS card_reward_rules (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    merchant_pattern TEXT NOT NULL,      -- e.g. "SWIGGY", "AMAZON", "FLIPKART", "BLINKIT", "FUEL"
    category_name TEXT,                  -- e.g. "Food & Dining", "E-Commerce", "Groceries", "Utilities"
    reward_percentage REAL NOT NULL,     -- e.g. 10.0, 5.0, 4.0, 2.0
    reward_description TEXT,             -- e.g. "10% cashback on Swiggy & Dineout"
    max_cap_per_month REAL,              -- e.g. 1500.0
    min_spend_per_txn REAL,              -- e.g. 100.0
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_card_reward_rules_account ON card_reward_rules(account_id);
CREATE INDEX IF NOT EXISTS idx_card_reward_rules_pattern ON card_reward_rules(merchant_pattern);

-- +goose Down
DROP INDEX IF EXISTS idx_card_reward_rules_pattern;
DROP INDEX IF EXISTS idx_card_reward_rules_account;
DROP TABLE IF EXISTS card_reward_rules;
ALTER TABLE accounts DROP COLUMN base_reward_rate;
ALTER TABLE accounts DROP COLUMN reward_type;
ALTER TABLE accounts DROP COLUMN card_color;
ALTER TABLE accounts DROP COLUMN payment_due_days;
ALTER TABLE accounts DROP COLUMN billing_day;
ALTER TABLE accounts DROP COLUMN fee_waiver_threshold;
ALTER TABLE accounts DROP COLUMN annual_fee;
