package db

import (
	"context"
	"testing"

	"github.com/leporo/sqlf"
	"gotest.tools/assert"
)

func TestGetRandomSQL(t *testing.T) {
	expected := `SELECT foo, moo FROM foo WHERE gid = ? LIMIT 1 OFFSET ABS(RANDOM()) % MAX((SELECT COUNT(*) FROM foo WHERE gid='bar'),1)`
	str := getRandom("foo", "bar").
		Select("foo").
		Select("moo").
		String()

	assert.Equal(t, expected, str, "get random SQL")
}

func TestRandomQuery(t *testing.T) {
	var err error
	err = Open(":memory:")
	defer Close()
	assert.NilError(t, err)

	tx, err := dbc.Begin()
	assert.NilError(t, err)
	defer tx.Commit()

	tx.Exec(`CREATE TABLE random_test ( data INT, gid VARCHAR(64) );`)

	_, err = sqlf.InsertInto("random_test").
		NewRow().Set("data", 2).Set("gid", "gid").
		NewRow().Set("data", 4).Set("gid", "gid").
		NewRow().Set("data", 6).Set("gid", "gid").
		Exec(context.TODO(), tx)
	assert.NilError(t, err)

	var res int64
	err = getRandom("random_test", "gid").
		Select("data").To(&res).
		QueryRow(context.TODO(), tx)
	assert.NilError(t, err)
	assert.Check(t, res == 2 || res == 4 || res == 6)
}
