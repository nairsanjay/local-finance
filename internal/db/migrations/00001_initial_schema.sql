-- +goose Up
CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    bank_name TEXT NOT NULL,
    account_type TEXT NOT NULL,
    account_number_mask TEXT,
    currency TEXT DEFAULT 'INR',
    opening_balance REAL DEFAULT 0.0,
    current_balance REAL DEFAULT 0.0,
    credit_limit REAL,
    billing_cycle_day INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS statement_imports (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    file_hash TEXT NOT NULL,
    statement_format TEXT NOT NULL,
    parser_used TEXT NOT NULL,
    start_date DATE,
    end_date DATE,
    total_transactions INTEGER,
    imported_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(account_id) REFERENCES accounts(id)
);

CREATE TABLE IF NOT EXISTS categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    parent_id TEXT,
    color_hex TEXT,
    icon TEXT,
    is_system BOOLEAN DEFAULT 0,
    FOREIGN KEY(parent_id) REFERENCES categories(id)
);

CREATE TABLE IF NOT EXISTS categorization_rules (
    id TEXT PRIMARY KEY,
    priority INTEGER DEFAULT 0,
    match_field TEXT NOT NULL,
    match_type TEXT NOT NULL,
    match_pattern TEXT NOT NULL,
    target_category_id TEXT NOT NULL,
    assign_tags TEXT,
    is_active BOOLEAN DEFAULT 1,
    FOREIGN KEY(target_category_id) REFERENCES categories(id)
);

CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    statement_import_id TEXT,
    tx_hash TEXT UNIQUE NOT NULL,
    tx_date DATE NOT NULL,
    value_date DATE,
    raw_narration TEXT NOT NULL,
    cleaned_payee TEXT,
    payment_mode TEXT,
    reference_number TEXT,
    tx_type TEXT NOT NULL,
    amount REAL NOT NULL,
    running_balance REAL,
    category_id TEXT,
    is_recurring BOOLEAN DEFAULT 0,
    notes TEXT,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(account_id) REFERENCES accounts(id),
    FOREIGN KEY(statement_import_id) REFERENCES statement_imports(id),
    FOREIGN KEY(category_id) REFERENCES categories(id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_tx_date ON transactions(tx_date);
CREATE INDEX IF NOT EXISTS idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);
CREATE INDEX IF NOT EXISTS idx_statement_imports_account_id ON statement_imports(account_id);

-- +goose Down
DROP INDEX IF EXISTS idx_statement_imports_account_id;
DROP INDEX IF EXISTS idx_transactions_category_id;
DROP INDEX IF EXISTS idx_transactions_tx_hash;
DROP INDEX IF EXISTS idx_transactions_tx_date;
DROP INDEX IF EXISTS idx_transactions_account_id;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS categorization_rules;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS statement_imports;
DROP TABLE IF EXISTS accounts;
