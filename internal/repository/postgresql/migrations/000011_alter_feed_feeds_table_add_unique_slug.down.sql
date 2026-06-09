-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

ALTER TABLE feed_feeds DROP CONSTRAINT feed_feeds_slug_unique;

-- Remove the UUID suffix appended during the up migration (dash + 8 hex characters).
UPDATE feed_feeds SET slug = LEFT(slug, LENGTH(slug) - 9);
