-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

ALTER TABLE bookmarks
ADD COLUMN fulltextsearch_tsv TSVECTOR;

CREATE INDEX idx_bookmarks_fulltextsearch_tsv
ON bookmarks
USING gin(fulltextsearch_tsv);

UPDATE bookmarks
SET fulltextsearch_tsv = to_tsvector(title) || ' ' || to_tsvector(replace(replace(description, '/', ' '), '.', ' ')) || to_tsvector(replace(array_to_string(tags, ' '), '/', ' '));
