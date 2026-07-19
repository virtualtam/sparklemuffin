-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

-- Backfill: append the first 8 hex characters of the UUID to existing slugs to ensure
-- uniqueness before enforcing the constraint.
UPDATE feed_feeds SET slug = slug || '-' || LEFT(uuid::text, 8);

ALTER TABLE feed_feeds ADD CONSTRAINT feed_feeds_slug_unique UNIQUE(slug);
