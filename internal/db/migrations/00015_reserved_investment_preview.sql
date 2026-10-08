-- +goose Up
-- This version was reserved before the investment preview shipped as 16/17.
-- Record it explicitly so fresh databases have a contiguous migration history.
-- It intentionally changes no schema or financial data.
SELECT 1;

-- +goose Down
SELECT 1;
