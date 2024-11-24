-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

CREATE TABLE IF NOT EXISTS users (
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    uuid                UUID NOT NULL PRIMARY KEY,
    email               TEXT UNIQUE NOT NULL,
    nick_name           TEXT UNIQUE NOT NULL,
    display_name        TEXT NOT NULL,
    password_hash       TEXT NOT NULL,
    is_admin            BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX idx_users_email
ON users ((LOWER(email)));
