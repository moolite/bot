CREATE TABLE channel_stats
( uid VARCHAR(128)
, points INT
, gid VARCHAR(128)
, PRIMARY KEY(uid)
, FOREIGN KEY(gid) REFERENCES groups)
