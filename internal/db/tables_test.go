package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/assert"
)

func TestTables(t *testing.T) {
	var err error

	dbc, err := sql.Open("sqlite3", ":memory:?cache=shared&mode=memory")
	assert.NilError(t, err)

	err = CreateTables(dbc)
	assert.NilError(t, err)
}
