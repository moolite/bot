package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/leporo/sqlf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"gotest.tools/assert"
)

func TestTables(t *testing.T) {
	ctx := context.Background()
	var err error

	dbc, err := sql.Open("sqlite3", ":memory:?cache=shared&mode=memory")
	assert.NilError(t, err)

	err = CreateTables(dbc)
	assert.NilError(t, err)

	gid := "one"
	text := "some text"
	data := "123456"
	kind := "photo"

	err = Query(ctx, dbc, InsertGroup(gid, "group name"))
	assert.ErrorType(t, err, sql.ErrNoRows)

	err = Query(ctx, dbc, InsertAbraxas(gid, text, kind))
	assert.ErrorType(t, err, sql.ErrNoRows)

	err = Query(ctx, dbc, InsertMedia(gid, data, kind, text))
	assert.ErrorType(t, err, sql.ErrNoRows)

	n := new(Media)

	err = sqlf.
		From(mediaTable).
		Select("kind", n.Kind).
		Select("description", n.Description).
		Select("data", n.Data).
		Select("gid", n.GID).
		Limit(1).
		QueryRow(ctx, dbc)
	assert.NilError(t, err)

	m := new(Media)
	stmt := SelectRandomMedia(gid, kind)
	err = QueryOne[Media](ctx, dbc, stmt, m)
	log.Debug().Msg(stmt.String())
	assert.NilError(t, err)

	assert.Equal(t, m.GID, gid)
	assert.Equal(t, m.Data, data)
	assert.Equal(t, m.Kind, kind)
	assert.Equal(t, m.Description, text)
}
