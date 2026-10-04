-- +goose Up

-- 1. Enhance Categories with default monthly budget target
ALTER TABLE categories ADD COLUMN monthly_budget REAL DEFAULT 0.0;

-- 2. Category Budgets table for month-specific budget targets and overrides
CREATE TABLE IF NOT EXISTS category_budgets (
    id TEXT PRIMARY KEY,
    category_id TEXT NOT NULL,
    month TEXT NOT NULL,               -- e.g. "2026-08" or "DEFAULT"
    monthly_limit REAL NOT NULL,       -- e.g. 12000.0
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE CASCADE,
    UNIQUE(category_id, month)
);

CREATE INDEX IF NOT EXISTS idx_category_budgets_cat_month ON category_budgets(category_id, month);
CREATE INDEX IF NOT EXISTS idx_category_budgets_month ON category_budgets(month);

-- +goose Down
DROP INDEX IF EXISTS idx_category_budgets_month;
DROP INDEX IF EXISTS idx_category_budgets_cat_month;
DROP TABLE IF EXISTS category_budgets;
ALTER TABLE categories DROP COLUMN monthly_budget;
