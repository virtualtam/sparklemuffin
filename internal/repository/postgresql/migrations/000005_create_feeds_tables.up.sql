-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

CREATE TABLE IF NOT EXISTS feed_feeds(
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    fetched_at    TIMESTAMPTZ,

    uuid          UUID        UNIQUE   NOT NULL PRIMARY KEY,
    feed_url      TEXT        UNIQUE   NOT NULL,
    title         TEXT        NOT NULL,
    slug          TEXT        NOT NULL,

    etag          TEXT        NOT NULL,
    last_modified TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS feed_entries(
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    uid          TEXT        UNIQUE   NOT NULL PRIMARY KEY,
    feed_uuid    UUID        NOT NULL,
    url          TEXT        NOT NULL,
    title        TEXT        NOT NULL,

    CONSTRAINT fk_feed FOREIGN KEY(feed_uuid) REFERENCES feed_feeds(uuid) ON DELETE CASCADE,
    CONSTRAINT unique_feed_url UNIQUE(feed_uuid, url)
);

CREATE TABLE IF NOT EXISTS feed_entries_metadata(
    user_uuid UUID    NOT NULL,
    entry_uid TEXT    NOT NULL,

    read      BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT fk_user FOREIGN KEY(user_uuid) REFERENCES users(uuid) ON DELETE CASCADE,
    CONSTRAINT fk_entry FOREIGN KEY(entry_uid) REFERENCES feed_entries(uid) ON DELETE CASCADE,
    CONSTRAINT pk_user_entry PRIMARY KEY(user_uuid, entry_uid)
);

CREATE TABLE IF NOT EXISTS feed_categories(
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    uuid       UUID        UNIQUE   NOT NULL PRIMARY KEY,
    user_uuid  UUID        NOT NULL,
    name       TEXT        NOT NULL,
    slug       TEXT        NOT NULL,

    CONSTRAINT fk_user FOREIGN KEY(user_uuid) REFERENCES users(uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS feed_subscriptions(
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    uuid          UUID        UNIQUE   NOT NULL PRIMARY KEY,
    category_uuid UUID        NOT NULL,
    feed_uuid     UUID        NOT NULL,
    user_uuid     UUID        NOT NULL,

    CONSTRAINT fk_category FOREIGN KEY(category_uuid) REFERENCES feed_categories(uuid) ON DELETE CASCADE,
    CONSTRAINT fk_feed FOREIGN KEY(feed_uuid) REFERENCES feed_feeds(uuid) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY(user_uuid) REFERENCES users(uuid) ON DELETE CASCADE
);
