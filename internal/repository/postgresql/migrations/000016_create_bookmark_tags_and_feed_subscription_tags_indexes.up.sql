-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

CREATE INDEX idx_bookmark_tags_user_uuid_tag_uuid ON bookmark_tags(user_uuid, tag_uuid);
CREATE INDEX idx_feed_subscription_tags_user_uuid_tag_uuid ON feed_subscription_tags(user_uuid, tag_uuid);
