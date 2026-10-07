-- +goose Up
-- Retain the published version for databases that already applied it.
-- Category-based expense calculations no longer require rewriting ledger flags.
SELECT 1;

-- +goose Down
-- Keep corrected flags: clearing them could undo user-confirmed transfers.
SELECT 1;
