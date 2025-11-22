-- Copyright (c) VirtualTam
-- SPDX-License-Identifier: MIT

ALTER TABLE feed_preferences
ADD COLUMN show_entry_summaries BOOLEAN NOT NULL DEFAULT TRUE;
