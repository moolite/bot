package db

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestTables(t *testing.T) {
	is := is.New(t)
	var err error

	err = Open(":memory:")
	defer Close()
	is.NoErr(err)

	err = Migrate()
	is.NoErr(err)

	var gid int64 = -123456
	text := "some text"
	data := "123456"
	kind := "photo"

	err = InsertGroup(context.TODO(), gid, "group name")
	is.NoErr(err)

	err = InsertAbraxas(context.TODO(), &Abraxas{GID: gid, Abraxas: "something", Kind: "photo"})
	is.NoErr(err)

	err = InsertMedia(context.TODO(), &Media{GID: gid, Data: data, Kind: kind, Description: text, Score: 0})
	is.NoErr(err)

	var c int64
	row := dbc.QueryRow(`SELECT COUNT(*) FROM media`)
	err = row.Scan(&c)
	is.NoErr(err)
	is.Equal(c, int64(1))

	n := &Media{}
	row = dbc.QueryRow(`SELECT kind,description,data,gid FROM media WHERE data=?`, data)
	err = row.Scan(&n.Kind, &n.Description, &n.Data, &n.GID)
	is.NoErr(err)
	is.Equal(n.GID, gid)

	m := &Media{GID: gid, Kind: kind}
	err = SelectRandomMedia(context.TODO(), m)
	is.NoErr(err)

	is.Equal(m.GID, gid)
	is.Equal(m.Data, data)
	is.Equal(m.Kind, kind)
	is.Equal(m.Description, text)

	mf := &MediaFts{}
	row = dbc.QueryRow(`SELECT rowid,description,gid FROM media_fts`)
	err = row.Scan(&mf.RowID, &mf.Description, &mf.GID)
	is.NoErr(err)

	is.Equal(mf.RowID, int64(1))
	is.Equal(mf.Description, m.Description)
	is.Equal(mf.GID, m.GID)

	s, err := SearchMedia(context.TODO(), gid, "some", 0)
	is.NoErr(err)
	is.Equal(len(s), 1)
	is.Equal(s[0].Description, text)

	s, err = SearchMedia(context.TODO(), gid, "nothing!", 0)
	is.NoErr(err)
	is.Equal(len(s), 0)

	err = MigrateDown()
	is.NoErr(err)
}
