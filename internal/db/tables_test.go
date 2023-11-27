package db

import (
	"context"
	"testing"

	"gotest.tools/assert"
)

func TestTables(t *testing.T) {
	var err error

	err = Open(":memory:")
	defer Close()
	assert.NilError(t, err)

	err = Migrate()
	assert.NilError(t, err)

	gid := "one"
	text := "some text"
	data := "123456"
	kind := "photo"

	err = InsertGroup(context.TODO(), gid, "group name")
	assert.NilError(t, err)

	err = InsertAbraxas(context.TODO(), &Abraxas{GID: gid, Abraxas: "something", Kind: "photo"})
	assert.NilError(t, err)

	err = InsertMedia(context.TODO(), &Media{gid, data, kind, text})
	assert.NilError(t, err)

	var c int64
	row := dbc.QueryRow(`SELECT COUNT(*) FROM media`)
	err = row.Scan(&c)
	assert.NilError(t, err)
	assert.Equal(t, c, int64(1))

	n := &Media{}
	row = dbc.QueryRow(`SELECT kind,description,data,gid FROM media WHERE data=?`, data)
	err = row.Scan(&n.Kind, &n.Description, &n.Data, &n.GID)
	assert.NilError(t, err)
	assert.Equal(t, n.GID, gid)
	assert.Equal(t, n.Data, data)

	m := &Media{GID: gid, Kind: kind}
	err = SelectRandomMedia(context.TODO(), m)
	assert.NilError(t, err)

	assert.Equal(t, m.GID, gid)
	assert.Equal(t, m.Data, data)
	assert.Equal(t, m.Kind, kind)
	assert.Equal(t, m.Description, text)

	err = MigrateDown()
	assert.NilError(t, err)
}
