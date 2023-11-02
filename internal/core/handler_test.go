package core

import (
	"fmt"
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestParseText(t *testing.T) {
	var res *BotRequest

	res = parseText("!cmd foo")
	assert.Assert(t, res.isCallout)

	res = parseText("/cmd foo")
	assert.Assert(t, res.isCommand)

	res = parseText("cmd foo")
	assert.Assert(t, res.isTrigger)

	res = parseText("/command@bot pupy so pupy")
	fmt.Fprintln(os.Stderr, "parsed :> ", res)

	assert.Assert(t, res.isCommand)
	assert.Equal(t, res.Abraxas, "command")
	assert.Equal(t, res.Rest, "pupy so pupy")

	res = parseText("!call me out baby!")
	assert.Assert(t, res.isCallout)
	assert.Equal(t, res.Abraxas, "call")
	assert.Equal(t, res.Rest, "me out baby!")

	res = parseText("trigger my pupy")
	assert.Assert(t, res.isTrigger)
	assert.Equal(t, res.Abraxas, "trigger")
	assert.Equal(t, res.Rest, "my pupy")
}
