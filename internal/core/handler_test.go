package core

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/telegram"
	"github.com/valyala/fastjson"
	"gotest.tools/assert"
)

func TestParseText(t *testing.T) {
	var res *BotRequest

	res = parseText("!cmd foo")
	assert.Equal(t, res.Kind, KindCallout)
	assert.Equal(t, res.Abraxas, "cmd")

	res = parseText("/cmd foo")
	assert.Equal(t, res.Kind, KindCommand)
	assert.Equal(t, res.Abraxas, "cmd")

	res = parseText("cmd foo")
	assert.Equal(t, res.Kind, KindTrigger)
	assert.Equal(t, res.Abraxas, "cmd")

	res = parseText("/command@bot pupy so pupy")
	assert.Equal(t, res.Kind, KindCommand)
	assert.Equal(t, res.Abraxas, "command")
	assert.Equal(t, res.Rest, "pupy so pupy")

	res = parseText("/backup")
	assert.Equal(t, res.Command, CmdBackup)

	res = parseText("/remember !callout text text")
	assert.Equal(t, res.Rest, "text text")
	assert.Equal(t, res.Command, CmdRemember)
	assert.Equal(t, res.Abraxas, "callout")

	res = parseText("!call me out baby!")
	assert.Equal(t, res.Kind, KindCallout)
	assert.Equal(t, res.Abraxas, "call")
	assert.Equal(t, res.Rest, "me out baby!")

	res = parseText("trigger my pupy")
	assert.Equal(t, res.Kind, KindTrigger)
	assert.Equal(t, res.Abraxas, "trigger")
	assert.Equal(t, res.Rest, "my pupy")
}

func tryHandle(t *testing.T, text string) (*telegram.WebhookResponse, error) {
	json, err := fastjson.Parse(text)
	if err != nil {
		t.Log("fastjson test error", err, "input:", text)
		return nil, err
	}
	return Handler(json)
}

func TestHandlerRemember(t *testing.T) {
	var err error
	gid := "123456"

	err = db.Open(":memory:")
	defer db.Close()
	assert.NilError(t, err, "error creating db")

	err = db.CreateTables()
	assert.NilError(t, err, "error creating tables")

	_, err = tryHandle(t, `{
		"chat": {"fid": 0 },
		"message": {}
	}`)
	assert.ErrorType(t, err, ErrParseNoChatID)

	_, err = tryHandle(t, `{
		"chat": {"fid": 0 }
	}`)
	assert.ErrorType(t, err, ErrParseNoMessage)

	err = db.InsertGroup(context.TODO(), gid, "title")
	assert.NilError(t, err)

	t.Run("empty photo", func(t *testing.T) {
		_, err := tryHandle(t, `{
			"message": {
				"chat": { "id": "`+gid+`" },
				"text": "/ricorda just a photo"
			}
		}`)
		expected := &telegram.WebhookResponse{}
		assert.NilError(t, err)
		assert.Equal(t, expected.ChatID, "")
	})

	tests := map[string]struct {
		gid         string
		kind        string
		description string
		fileId      string
	}{
		"photo": {
			kind:        "photo",
			gid:         gid,
			description: "a photo",
			fileId:      "AAAAAAAA",
		},
		"animation": {
			kind:        "animation",
			gid:         gid,
			description: "a new animation",
			fileId:      "BBBBBB",
		},
		"video": {
			kind:        "video",
			gid:         gid,
			description: "a new video",
			fileId:      "CCCCCC",
		},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			var err error
			var res *telegram.WebhookResponse

			res, err = tryHandle(t, `{
				"message": {
					"chat": { "id": "`+test.gid+`" },
					"text": "/ricorda `+test.description+`",
					"`+test.kind+`": [{"file_id":"`+test.fileId+`"},
									  {"file_id":"wrong"}]
				}
			}`)
			assert.NilError(t, err)
			assert.Check(t, res != nil, "should return a WebhookResponse")
			assert.Check(t, *res.Text == textRemember, "should return human acknowledgement")

			m := &db.Media{
				GID:  gid,
				Data: test.fileId,
			}
			err = db.SelectOneMediaByData(context.TODO(), m)
			assert.NilError(t, err)
			assert.Equal(t, m.Description, test.description)
			assert.Equal(t, m.Kind, test.kind)

		})
	}
}

func TestHandlerCallout(t *testing.T) {
	var err error
	var res *telegram.WebhookResponse
	gid := "123456"

	err = db.Open(":memory:")
	defer db.Close()
	assert.NilError(t, err, "error creating db")

	err = db.CreateTables()
	assert.NilError(t, err, "error creating tables")

	tests := map[string]struct {
		remember string
		callout  string
		message  string
		text     string
		expected string
	}{
		"simple": {
			remember: "/remember !calltext text",
			callout:  "calltext",
			message:  "!calltext foo",
			text:     "text",
			expected: "text",
		},
		"emoji": {
			remember: "/remember !callmoji text 🐶🎉🥳",
			callout:  "call",
			message:  "!callmoji foo",
			text:     "text 🐶🎉🥳",
			expected: "text 🐶🎉🥳",
		},
		"text with symbols": {
			remember: "/remember !marrano %, you are very marrano!",
			callout:  "marrano",
			message:  "!callmoji foo",
			text:     "%, you are very marrano!",
			expected: "foo, you are very marrano!",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			res, err = tryHandle(t, `{
				"message": {
					"chat": {"id":"`+gid+`"},
					"text": "`+test.remember+`"
				}
			}`)
			assert.NilError(t, err)
			assert.Equal(t, res.Text, textRemember)

			c := &db.Callout{
				GID:     gid,
				Callout: test.callout,
			}
			err := db.SelectOneCallout(context.TODO(), c)
			assert.NilError(t, err)
			assert.Equal(t, c.Text, test.text)

			res, err = tryHandle(t, `{
				"message": {
					"chat": { "id": "`+gid+`" },
					"text": "`+test.message+`"
			}`)
			assert.NilError(t, err)
			assert.Equal(t, res.Text, test.expected)
		})
	}
}
