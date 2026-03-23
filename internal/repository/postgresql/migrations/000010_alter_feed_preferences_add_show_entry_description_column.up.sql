-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

ALTER TABLE feed_preferences
ADD COLUMN show_entry_summaries BOOLEAN NOT NULL DEFAULT TRUE;
