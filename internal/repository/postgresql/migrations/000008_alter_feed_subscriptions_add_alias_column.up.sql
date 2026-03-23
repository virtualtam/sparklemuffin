-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

ALTER TABLE feed_subscriptions
ADD COLUMN alias TEXT NOT NULL DEFAULT '';
