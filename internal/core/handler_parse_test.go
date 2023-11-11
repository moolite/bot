package core

import (
	"testing"

	"gotest.tools/assert"
)

func TestParseText(t *testing.T) {
	var req *BotRequest

	req = parseText("!cmd foo")
	assert.Equal(t, req.Kind, KindCallout)
	assert.Equal(t, req.Abraxas, "cmd")

	req = parseText("/cmd foo")
	assert.Equal(t, req.Kind, KindCommand)
	assert.Equal(t, req.Abraxas, "cmd")

	req = parseText("cmd foo")
	assert.Equal(t, req.Kind, KindTrigger)
	assert.Equal(t, req.Abraxas, "cmd")

	req = parseText("/command@bot pupy so pupy")
	assert.Equal(t, req.Kind, KindCommand)
	assert.Equal(t, req.Abraxas, "command")
	assert.Equal(t, req.Rest, "pupy so pupy")

	req = parseText("/backup")
	assert.Equal(t, req.Command, CmdBackup)

	req = parseText("/remember !callout text text")
	assert.Equal(t, req.Rest, "!callout text text")
	assert.Equal(t, req.Command, CmdRemember)
	assert.Equal(t, req.RememberKind, KindCalloutCmd)
	assert.Equal(t, req.RememberAbraxas, "callout")

	req = parseText("!call me out baby!")
	assert.Equal(t, req.Kind, KindCallout)
	assert.Equal(t, req.Abraxas, "call")
	assert.Equal(t, req.Rest, "me out baby!")

	req = parseText("trigger my pupy")
	assert.Equal(t, req.Kind, KindTrigger)
	assert.Equal(t, req.Abraxas, "trigger")
	assert.Equal(t, req.Rest, "my pupy")
}
