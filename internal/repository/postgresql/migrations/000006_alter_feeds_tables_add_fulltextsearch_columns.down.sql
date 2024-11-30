-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

-- Feeds
ALTER TABLE feed_feeds
DROP COLUMN description,
DROP COLUMN fulltextsearch_tsv;

-- Feed Entries
DROP INDEX IF EXISTS idx_feed_entries_fulltextsearch_tsv;

ALTER TABLE feed_entries
DROP COLUMN summary,
DROP COLUMN fulltextsearch_tsv;
