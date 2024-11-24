-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

DROP INDEX IF EXISTS idx_bookmarks_fulltextsearch_tsv;

ALTER TABLE bookmarks
DROP COLUMN fulltextsearch_tsv;
