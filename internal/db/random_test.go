package db

import (
	"testing"
)

func TestGetRandom(t *testing.T) {
	expected := `SELECT foo, moo FROM foo WHERE gid = ? LIMIT ? OFFSET ? OFFSET ABS(RANDOM()) % MAX((SELECT COUNT(*) FROM ? WHERE gid = ?), 1)`
	str := getRandom("foo", "bar").
		Select("foo").
		Select("moo").
		String()

	if str != expected {
		t.Error("SQL error", str)
	}
}
