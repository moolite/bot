CREATE VIRTUAL TABLE media_fts USING
fts5(
, description
, gid UNINDEXED
);


CREATE TRIGGER media_oninsert AFTER INSERT ON media BEGIN
    INSERT INTO media_fts(rowid, description, gid)
    VALUES (new.rowid, new.description, new.gid);
END;

CREATE TRIGGER media_onupdate AFTER UPDATE ON media BEGIN
    UPDATE media_fts
    SET description = new.description
    WHERE rowid = new.rowid;
END;

CREATE TRIGGER media_ondelete AFTER DELETE ON media BEGIN
  DELETE FROM media_fts WHERE rowid = old.rowid;
END;
