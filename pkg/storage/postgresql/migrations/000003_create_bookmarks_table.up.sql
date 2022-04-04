CREATE TABLE IF NOT EXISTS bookmarks(
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    uid         TEXT UNIQUE NOT NULL PRIMARY KEY,
    user_uuid   UUID NOT NULL,
    url         TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT DEFAULT '',

    CONSTRAINT fk_user FOREIGN KEY(user_uuid) REFERENCES users(uuid) ON DELETE CASCADE
);
