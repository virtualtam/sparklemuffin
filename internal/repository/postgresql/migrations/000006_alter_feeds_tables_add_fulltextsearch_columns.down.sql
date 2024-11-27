-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

-- Feeds
ALTER TABLE feed_feeds
DROP COLUMN description;

-- Feed Entries
DROP INDEX IF EXISTS idx_feed_entries_fulltextsearch_tsv;

ALTER TABLE feed_entries
DROP COLUMN fulltextsearch_tsv;
