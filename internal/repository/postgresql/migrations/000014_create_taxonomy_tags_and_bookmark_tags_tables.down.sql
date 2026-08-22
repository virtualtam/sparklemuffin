-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

ALTER TABLE bookmarks
ADD COLUMN tags TEXT[];

DROP TABLE IF EXISTS bookmark_tags;

ALTER TABLE bookmarks
DROP CONSTRAINT IF EXISTS unique_user_uuid_uid;

DROP TABLE IF EXISTS taxonomy_tags;
