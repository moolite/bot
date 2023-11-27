CREATE TABLE IF NOT EXISTS media
( data        VARCHAR(512) NOT NULL
, kind        VARCHAR(64)  NOT NULL
, gid         VARCHAR(64)  NOT NULL
, description TEXT
, PRIMARY KEY(data,gid)
, FOREIGN KEY(gid) REFERENCES groups(gid)
);
