-- +goose Up

-- 1. Create Subscriptions and Recurring Mandates Table
CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    merchant_pattern TEXT NOT NULL,
    category_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
    account_id TEXT REFERENCES accounts(id) ON DELETE SET NULL,
    frequency TEXT NOT NULL DEFAULT 'MONTHLY', -- 'MONTHLY', 'QUARTERLY', 'YEARLY', 'WEEKLY'
    expected_amount REAL NOT NULL,
    currency TEXT NOT NULL DEFAULT 'INR',
    billing_day INTEGER,                      -- Day of month (e.g. 5, 16, 27)
    next_due_date TEXT,                       -- YYYY-MM-DD
    last_paid_date TEXT,                      -- YYYY-MM-DD
    last_paid_amount REAL,
    status TEXT NOT NULL DEFAULT 'ACTIVE',    -- 'ACTIVE', 'PAUSED', 'CANCELLED'
    is_auto_detected INTEGER NOT NULL DEFAULT 1,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_next_due ON subscriptions(next_due_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_category ON subscriptions(category_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_account ON subscriptions(account_id);

-- +goose Down
DROP INDEX IF EXISTS idx_subscriptions_account;
DROP INDEX IF EXISTS idx_subscriptions_category;
DROP INDEX IF EXISTS idx_subscriptions_next_due;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP TABLE IF EXISTS subscriptions;
