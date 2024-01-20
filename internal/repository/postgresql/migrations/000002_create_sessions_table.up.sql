-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

CREATE TABLE IF NOT EXISTS sessions (
    remember_token_hash       TEXT UNIQUE NOT NULL PRIMARY KEY,
    remember_token_expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_uuid                 UUID NOT NULL,

    CONSTRAINT fk_user FOREIGN KEY(user_uuid) REFERENCES users(uuid) ON DELETE CASCADE
);
