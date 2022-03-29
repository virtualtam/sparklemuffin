CREATE TABLE IF NOT EXISTS users (
    uuid                UUID NOT NULL PRIMARY KEY,
    email               TEXT UNIQUE NOT NULL,
    password_hash       TEXT NOT NULL DEFAULT '',
    remember_token_hash TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX idx_users_email
ON users ((lower(email)));
