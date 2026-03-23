-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

ALTER TABLE feed_feeds
ADD COLUMN hash_xxhash64 BIGINT NOT NULL DEFAULT 0;
