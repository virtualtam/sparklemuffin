CREATE TABLE IF NOT EXISTS users (
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    uuid                UUID NOT NULL PRIMARY KEY,
    email               TEXT UNIQUE NOT NULL,
    password_hash       TEXT NOT NULL DEFAULT '',
    remember_token_hash TEXT NOT NULL DEFAULT '',
    is_admin            BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX idx_users_email
ON users ((lower(email)));
