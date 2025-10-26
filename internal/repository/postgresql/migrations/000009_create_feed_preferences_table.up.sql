-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

CREATE TYPE feed_entry_visibility AS ENUM(
    'ALL',
    'READ',
    'UNREAD'
);

CREATE TABLE IF NOT EXISTS feed_preferences(
    updated_at   TIMESTAMPTZ           NOT NULL DEFAULT NOW(),

    user_uuid    UUID                  UNIQUE   NOT NULL PRIMARY KEY,
    show_entries feed_entry_visibility NOT NULL DEFAULT 'ALL'::feed_entry_visibility,

    CONSTRAINT fk_user FOREIGN KEY(user_uuid) REFERENCES users(uuid) ON DELETE CASCADE
);

INSERT INTO feed_preferences(user_uuid)
SELECT uuid
FROM users;
