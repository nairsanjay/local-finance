-- +goose Up
CREATE TABLE investment_snapshots (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    account_ref TEXT NOT NULL,
    as_of TEXT NOT NULL,
    file_hash TEXT NOT NULL UNIQUE,
    imported_at TEXT NOT NULL,
    data_json TEXT NOT NULL
);
CREATE INDEX idx_investment_portfolio_date ON investment_snapshots(provider, account_ref, as_of);

-- +goose Down
DROP TABLE investment_snapshots;
