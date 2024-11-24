-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

ALTER TABLE feed_entries
ADD COLUMN fulltextsearch_tsv TSVECTOR;

CREATE INDEX idx_feed_entries_fulltextsearch_tsv
ON feed_entries
USING gin(fulltextsearch_tsv);

UPDATE feed_entries
SET fulltextsearch_tsv = to_tsvector(title);
