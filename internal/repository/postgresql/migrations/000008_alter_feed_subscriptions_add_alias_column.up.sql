-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

ALTER TABLE feed_subscriptions
ADD COLUMN alias TEXT NOT NULL DEFAULT '';
