-- +goose Up

-- 1. Enhance Accounts with card network and card variant
ALTER TABLE accounts ADD COLUMN card_network TEXT;
ALTER TABLE accounts ADD COLUMN card_variant TEXT;

-- 2. Enhance Transactions with merchant category and cashback/reward points
ALTER TABLE transactions ADD COLUMN merchant_category TEXT;
ALTER TABLE transactions ADD COLUMN cashback_amount REAL DEFAULT 0.0;
ALTER TABLE transactions ADD COLUMN reward_points_earned REAL DEFAULT 0.0;

-- 3. Enhance Credit Card Bills with cashback, credit limit & finance charges breakdown
ALTER TABLE credit_card_bills ADD COLUMN cashback_earned REAL DEFAULT 0.0;
ALTER TABLE credit_card_bills ADD COLUMN cashback_credited REAL DEFAULT 0.0;
ALTER TABLE credit_card_bills ADD COLUMN finance_charges REAL DEFAULT 0.0;
ALTER TABLE credit_card_bills ADD COLUMN credit_limit REAL;
ALTER TABLE credit_card_bills ADD COLUMN available_credit_limit REAL;

CREATE INDEX IF NOT EXISTS idx_transactions_merchant_category ON transactions(merchant_category);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_merchant_category;
ALTER TABLE credit_card_bills DROP COLUMN available_credit_limit;
ALTER TABLE credit_card_bills DROP COLUMN credit_limit;
ALTER TABLE credit_card_bills DROP COLUMN finance_charges;
ALTER TABLE credit_card_bills DROP COLUMN cashback_credited;
ALTER TABLE credit_card_bills DROP COLUMN cashback_earned;
ALTER TABLE transactions DROP COLUMN reward_points_earned;
ALTER TABLE transactions DROP COLUMN cashback_amount;
ALTER TABLE transactions DROP COLUMN merchant_category;
ALTER TABLE accounts DROP COLUMN card_variant;
ALTER TABLE accounts DROP COLUMN card_network;
