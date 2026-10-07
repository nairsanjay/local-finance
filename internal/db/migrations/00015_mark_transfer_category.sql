-- +goose Up
-- Existing Self / P2P rows should use the transfer exclusion already applied
-- by income and expense calculations. Match the stable category ID.
UPDATE transactions
SET is_transfer = 1
WHERE category_id = 'cat_transfers' AND is_transfer = 0;

-- +goose Down
-- Keep corrected flags: clearing them could undo user-confirmed transfers.
SELECT 1;
