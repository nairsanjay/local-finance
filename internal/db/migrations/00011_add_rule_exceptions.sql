-- +goose Up
ALTER TABLE categorization_rules ADD COLUMN exclude_pattern TEXT DEFAULT '';
ALTER TABLE categorization_rules ADD COLUMN tx_type TEXT DEFAULT 'ALL';

UPDATE categorization_rules
SET tx_type = 'CREDIT',
    exclude_pattern = 'maid, driver, cook, helper, staff, advance'
WHERE id = 'rule_salary';

-- +goose Down
ALTER TABLE categorization_rules DROP COLUMN exclude_pattern;
ALTER TABLE categorization_rules DROP COLUMN tx_type;
