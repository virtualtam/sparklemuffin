-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

DROP INDEX IF EXISTS idx_feed_subscriptions_fulltextsearch_tsv;

ALTER TABLE feed_subscriptions
DROP COLUMN fulltextsearch_tsv;
