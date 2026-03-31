package db

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
)

func TestTables(t *testing.T) {
	is := is.New(t)

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	err := InsertGroup(context.TODO(), -123456, "group name")
	is.NoErr(err)

	err = InsertAbraxas(context.TODO(), &Abraxas{GID: -123456, Abraxas: "something", Kind: "photo"})
	is.NoErr(err)

	err = InsertMedia(context.TODO(), &Media{GID: -123456, Data: "123456", Kind: "photo", Description: "some text", Score: 0})
	is.NoErr(err)

	db := Default().DB()

	var c int64
	row := db.QueryRow(`SELECT COUNT(*) FROM media`)
	err = row.Scan(&c)
	is.NoErr(err)
	is.Equal(c, int64(1))

	n := &Media{}
	row = db.QueryRow(`SELECT kind, description, data, gid FROM media WHERE data=?`, "123456")
	err = row.Scan(&n.Kind, &n.Description, &n.Data, &n.GID)
	is.NoErr(err)
	is.Equal(n.GID, int64(-123456))
	is.Equal(n.Data, "123456")
	is.Equal(n.Description, "some text")
	is.Equal(n.Kind, "photo")

	m := &Media{GID: -123456, Kind: "photo"}
	err = SelectRandomMedia(context.TODO(), m)
	is.NoErr(err)
	is.Equal(m.GID, int64(-123456))
	is.Equal(m.Data, "123456")
	is.Equal(m.Kind, "photo")
	is.Equal(m.Description, "some text")
}

func TestTablesWithFTS(t *testing.T) {
	is := is.New(t)

	dbPath := filepath.Join(t.TempDir(), "test.db")
	is.NoErr(Open(dbPath))
	is.NoErr(Migrate())
	defer Close()

	err := InsertGroup(context.TODO(), -123456, "group name")
	is.NoErr(err)

	err = InsertMedia(context.TODO(), &Media{GID: -123456, Data: "123456", Kind: "photo", Description: "some text", Score: 0})
	is.NoErr(err)

	db := Default().DB()

	var mf MediaFts
	row := db.QueryRow(`SELECT rowid, description, gid FROM media_fts`)
	err = row.Scan(&mf.RowID, &mf.Description, &mf.GID)
	is.NoErr(err)
	is.Equal(mf.GID, int64(-123456))
	is.Equal(mf.Description, "some text")
	is.Equal(mf.RowID, int64(1))

	s, err := SearchMedia(context.TODO(), -123456, "some", 0)
	is.NoErr(err)
	is.Equal(len(s), 1)

	s, err = SearchMedia(context.TODO(), -123456, "nothing!", 0)
	is.NoErr(err)
	is.Equal(len(s), 0)
}
