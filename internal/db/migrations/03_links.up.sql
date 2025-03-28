CREATE TABLE IF NOT EXISTS links
( url VARCHAR(256) NOT NULL
, text TEXT
, gid VARCHAR(64) NOT NULL
, PRIMARY KEY(url, gid)
, FOREIGN KEY(gid) REFERENCES groups
);
