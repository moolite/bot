CREATE TABLE IF NOT EXISTS callouts
( callout VARCHAR(128) NOT NULL
, gid     VARCHAR(64)  NOT NULL
, text    TEXT
, PRIMARY KEY(callout,gid)
, FOREIGN KEY(gid) REFERENCES groups
);
