DROP INDEX IF EXISTS idx_bookmarks_fulltextsearch_tsv;

ALTER TABLE bookmarks
DROP column fulltextsearch_tsv,
