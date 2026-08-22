-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

ALTER TABLE feed_subscriptions
ADD COLUMN tags TEXT[];

DROP TABLE IF EXISTS feed_subscription_tags;

ALTER TABLE feed_subscriptions
DROP CONSTRAINT IF EXISTS unique_user_uuid_uuid;
