-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

CREATE TABLE IF NOT EXISTS taxonomy_tags(
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    tag_uuid     UUID        UNIQUE   NOT NULL PRIMARY KEY,
    user_uuid    UUID        NOT NULL,
    tag_name     TEXT        NOT NULL,
    tag_name_tsv TSVECTOR,

    CONSTRAINT fk_user FOREIGN KEY(user_uuid) REFERENCES users(uuid) ON DELETE CASCADE,
    CONSTRAINT unique_user_uuid_tag_name UNIQUE(user_uuid, tag_name),
    CONSTRAINT unique_user_uuid_tag_uuid UNIQUE(user_uuid, tag_uuid)
);

CREATE INDEX idx_taxonomy_tags_tag_name_tsv
ON taxonomy_tags
USING gin(tag_name_tsv);

-- Normalize bookmark tags.
ALTER TABLE bookmarks
ADD CONSTRAINT unique_user_uuid_uid UNIQUE(user_uuid, uid);

CREATE TABLE IF NOT EXISTS bookmark_tags(
    user_uuid    UUID NOT NULL,
    bookmark_uid TEXT NOT NULL,
    tag_uuid     UUID NOT NULL,

    CONSTRAINT fk_bookmark FOREIGN KEY(user_uuid, bookmark_uid) REFERENCES bookmarks(user_uuid, uid) ON DELETE CASCADE,
    CONSTRAINT fk_tag FOREIGN KEY(user_uuid, tag_uuid) REFERENCES taxonomy_tags(user_uuid, tag_uuid) ON DELETE CASCADE,
    CONSTRAINT pk_bookmark_tag PRIMARY KEY(bookmark_uid, tag_uuid)
);

-- Backfill taxonomy_tags from the existing bookmarks.tags arrays.
INSERT INTO taxonomy_tags(tag_uuid, user_uuid, tag_name, tag_name_tsv)
SELECT
    GEN_RANDOM_UUID() AS generated_uuid,
    user_uuid,
    tag_name,
    TO_TSVECTOR(tag_name) AS tag_name_tsv
FROM(
    SELECT DISTINCT
        user_uuid,
        UNNEST(tags) AS tag_name
    FROM bookmarks
) AS distinct_bookmark_tags
ON CONFLICT(user_uuid, tag_name) DO NOTHING;

-- Backfill bookmark_tags from the existing bookmarks.tags arrays.
INSERT INTO bookmark_tags(user_uuid, bookmark_uid, tag_uuid)
SELECT
    b.user_uuid,
    b.uid,
    tt.tag_uuid
FROM(
    SELECT DISTINCT
        user_uuid,
        uid,
        UNNEST(tags) AS tag_name
    FROM bookmarks
) AS b
INNER JOIN taxonomy_tags AS tt ON b.user_uuid = tt.user_uuid AND b.tag_name = tt.tag_name
ON CONFLICT(bookmark_uid, tag_uuid) DO NOTHING;

-- Bookmark tags are now normalized: drop the legacy array column.
ALTER TABLE bookmarks
DROP COLUMN tags;
