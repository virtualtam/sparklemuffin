-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

-- Feeds
ALTER TABLE feed_feeds
ADD COLUMN description        TEXT NOT NULL DEFAULT '',
ADD COLUMN fulltextsearch_tsv TSVECTOR;

CREATE INDEX idx_feed_feeds_fulltextsearch_tsv
ON feed_feeds
USING gin(fulltextsearch_tsv);

UPDATE feed_feeds
SET fulltextsearch_tsv = to_tsvector(replace(title, '.', ' ') || ' ' || replace(description, '.', ' '));

-- Feed Entries
ALTER TABLE feed_entries
ADD COLUMN summary            TEXT NOT NULL DEFAULT '',
ADD COLUMN textrank_terms     TEXT[],
ADD COLUMN fulltextsearch_tsv TSVECTOR;

CREATE INDEX idx_feed_entries_fulltextsearch_tsv
ON feed_entries
USING gin(fulltextsearch_tsv);

UPDATE feed_entries
SET fulltextsearch_tsv = to_tsvector(title) || to_tsvector(array_to_string(textrank_terms, ' '));
