package db

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetRandomSQL(t *testing.T) {
	expected := `SELECT foo, moo FROM foo WHERE gid = ? LIMIT ? OFFSET ABS(RANDOM()) % MAX((SELECT COUNT(*) FROM ? WHERE gid = ?), 1)`
	str := getRandom("foo", "bar").
		Select("foo").
		Select("moo").
		String()

	assert.Equal(t, expected, str, "get random SQL")
}
