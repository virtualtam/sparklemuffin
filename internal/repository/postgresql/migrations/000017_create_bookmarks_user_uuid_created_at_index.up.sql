-- Copyright VirtualTam 2022, 2026
-- SPDX-License-Identifier: MIT

CREATE INDEX idx_bookmarks_user_uuid_created_at ON bookmarks(user_uuid, created_at);
