CREATE TABLE aliases
( name VARCHAR(128) NOT NULL
, target VARCHAR(128) NOT NULL
, PRIMARY KEY(name, target)
);
