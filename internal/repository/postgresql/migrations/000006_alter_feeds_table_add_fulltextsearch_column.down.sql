-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

DROP INDEX IF EXISTS idx_feed_entries_fulltextsearch_tsv;

ALTER TABLE feed_entries
DROP COLUMN fulltextsearch_tsv;
