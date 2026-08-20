-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

ALTER TABLE feed_subscriptions
ADD COLUMN fulltextsearch_tsv TSVECTOR;

CREATE INDEX idx_feed_subscriptions_fulltextsearch_tsv
ON feed_subscriptions
USING gin(fulltextsearch_tsv);

UPDATE feed_subscriptions
SET fulltextsearch_tsv = to_tsvector(replace(replace(alias, '/', ' '), '.', ' ')) || to_tsvector(replace(replace(array_to_string(tags, ' '), '/', ' '), '.', ' '));
