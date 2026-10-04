-- +goose Up
CREATE TABLE IF NOT EXISTS app_security (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    auth_enabled BOOLEAN NOT NULL DEFAULT 0,
    password_hash TEXT NOT NULL DEFAULT '',
    auto_lock_minutes INTEGER NOT NULL DEFAULT 60,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO app_security (id, auth_enabled, password_hash, auto_lock_minutes)
VALUES (1, 0, '', 60)
ON CONFLICT(id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS app_security;
