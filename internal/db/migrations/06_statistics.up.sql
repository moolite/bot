CREATE TABLE IF NOT EXISTS statistics_kind
( kind_id INTEGER PRIMARY KEY ASC
, name      VARCHAR(16)
, type      VARCHAR(16)
, trigger   VARCHAR(128)
, gid       VARCHAR(64)
, is_regexp BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS statistics
( value   INTEGER
, date    TIMESTAMP
, kind_id INTEGER
, FOREIGN KEY(kind_id) REFERENCES statistics_kind
, PRIMARY KEY(date,kind_id)
);

CREATE TRIGGER "statistics_date"
AFTER INSERT ON "statistics"
BEGIN
UPDATE statistics
SET date = datetime('now')
WHERE rowid = NEW.rowid
;
END;
