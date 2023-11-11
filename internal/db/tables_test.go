package db

import (
	"context"
	"testing"

	"github.com/leporo/sqlf"
	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/assert"
)

func TestTables(t *testing.T) {
	var err error

	err = Open(":memory:")
	defer Close()
	assert.NilError(t, err)

	err = CreateTables()
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
	err = sqlf.From(mediaTable).
		Select("COUNT(*)").To(&c).
		QueryRowAndClose(context.TODO(), dbc)
	assert.NilError(t, err)
	assert.Equal(t, c, int64(1))

	n := &Media{}

	err = sqlf.
		From(mediaTable).
		Select("kind").To(&n.Kind).
		Select("description").To(&n.Description).
		Select("data").To(&n.Data).
		Select("gid").To(&n.GID).
		Where("data = ?", data).
		Limit(1).
		QueryRow(context.TODO(), dbc)
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
}
