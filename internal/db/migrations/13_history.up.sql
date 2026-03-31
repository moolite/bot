CREATE TABLE IF NOT EXISTS history (
    rowid       INTEGER PRIMARY KEY AUTOINCREMENT,
    gid         VARCHAR(64) NOT NULL REFERENCES groups(gid),
    message_id  INTEGER NOT NULL,
    user_id     INTEGER,
    username    TEXT,
    text        TEXT NOT NULL,
    timestamp   INTEGER NOT NULL,
    embedding   BLOB,
    UNIQUE(gid, message_id)
);

CREATE INDEX IF NOT EXISTS idx_history_gid ON history(gid);
CREATE INDEX IF NOT EXISTS idx_history_timestamp ON history(timestamp);
