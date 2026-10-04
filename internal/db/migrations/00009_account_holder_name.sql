-- +goose Up
ALTER TABLE accounts ADD COLUMN account_holder_name TEXT;

-- +goose Down
ALTER TABLE accounts DROP COLUMN account_holder_name;
