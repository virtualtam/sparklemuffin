-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

-- Normalize feed subscription tags.
ALTER TABLE feed_subscriptions
ADD CONSTRAINT unique_user_uuid_uuid UNIQUE(user_uuid, uuid);

CREATE TABLE IF NOT EXISTS feed_subscription_tags(
    user_uuid         UUID NOT NULL,
    subscription_uuid UUID NOT NULL,
    tag_uuid          UUID NOT NULL,

    CONSTRAINT fk_subscription FOREIGN KEY(user_uuid, subscription_uuid) REFERENCES feed_subscriptions(user_uuid, uuid) ON DELETE CASCADE,
    CONSTRAINT fk_tag FOREIGN KEY(user_uuid, tag_uuid) REFERENCES taxonomy_tags(user_uuid, tag_uuid) ON DELETE CASCADE,
    CONSTRAINT pk_subscription_tag PRIMARY KEY(subscription_uuid, tag_uuid)
);

-- Backfill taxonomy_tags from the existing feed_subscriptions.tags arrays.
INSERT INTO taxonomy_tags(tag_uuid, user_uuid, tag_name)
SELECT
    GEN_RANDOM_UUID() AS generated_uuid,
    user_uuid,
    tag_name
FROM(
    SELECT DISTINCT
        user_uuid,
        UNNEST(tags) AS tag_name
    FROM feed_subscriptions
) AS distinct_subscription_tags
ON CONFLICT(user_uuid, tag_name) DO NOTHING;

-- Backfill feed_subscription_tags from the existing feed_subscriptions.tags arrays.
INSERT INTO feed_subscription_tags(user_uuid, subscription_uuid, tag_uuid)
SELECT
    s.user_uuid,
    s.uuid,
    tt.tag_uuid
FROM(
    SELECT DISTINCT
        user_uuid,
        uuid,
        UNNEST(tags) AS tag_name
    FROM feed_subscriptions
) AS s
INNER JOIN taxonomy_tags AS tt ON s.user_uuid = tt.user_uuid AND s.tag_name = tt.tag_name
ON CONFLICT(subscription_uuid, tag_uuid) DO NOTHING;

-- Feed subscription tags are now normalized: drop the legacy array column.
ALTER TABLE feed_subscriptions
DROP COLUMN tags;
