CREATE TABLE channel_stats
( ts TIMESTAMP
, uid VARCHAR(128)
, points INT
, gid VARCHAR(128)
, PRIMARY KEY(uid,ts)
, FOREIGN KEY(gid) REFERENCES groups)
