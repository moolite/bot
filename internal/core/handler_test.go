package core

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/moolite/bot/internal/db"
	"gotest.tools/assert"
)

func TestParseText(t *testing.T) {
	var res *BotRequest

	res = parseText("!cmd foo")
	assert.Equal(t, res.Kind, KindCallout)

	res = parseText("/cmd foo")
	assert.Equal(t, res.Kind, KindCommand)

	res = parseText("cmd foo")
	assert.Equal(t, res.Kind, KindTrigger)

	res = parseText("/command@bot pupy so pupy")
	assert.Equal(t, res.Kind, KindCommand)
	assert.Equal(t, res.Abraxas, "command")
	assert.Equal(t, res.Rest, "pupy so pupy")

	res = parseText("!call me out baby!")
	assert.Equal(t, res.Kind, KindCallout)
	assert.Equal(t, res.Abraxas, "call")
	assert.Equal(t, res.Rest, "me out baby!")

	res = parseText("trigger my pupy")
	assert.Equal(t, res.Kind, KindTrigger)
	assert.Equal(t, res.Abraxas, "trigger")
	assert.Equal(t, res.Rest, "my pupy")
}

func TestHandler(t *testing.T) {
	dbc, err := sql.Open("sqlite3", ":memory:?cache=shared&mode=memory")
	assert.NilError(t, err, "error creating db")

	err = db.CreateTables(dbc)
	assert.NilError(t, err, "error creating tables")
}
