CREATE TABLE channel_stats
( uid VARCHAR(128)
, points INT
, gid VARCHAR(64)
, PRIMARY KEY(uid, gid)
, FOREIGN KEY(gid) REFERENCES groups)
