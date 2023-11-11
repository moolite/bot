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

func prepareDB(t *testing.T, gid string) {
	var err error

	err = db.Open(":memory:")
	assert.NilError(t, err, "error creating db")

	err = db.CreateTables()
	assert.NilError(t, err, "error creating tables")

	err = db.InsertGroup(context.TODO(), gid, "title")
	assert.NilError(t, err)
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

	prepareDB(t, gid)
	defer db.Close()

	_, err = tryHandle(t, `{
		"chat": {"fid": 0 },
		"message": {}
	}`)
	assert.ErrorType(t, err, ErrParseNoChatID)

	_, err = tryHandle(t, `{
		"chat": {"fid": 0 }
	}`)
	assert.ErrorType(t, err, ErrParseNoMessage)

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
			kind:        "video",
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
	gid := "123456"

	prepareDB(t, gid)
	defer db.Close()

	tests := map[string]struct {
		remember string
		callout  string
		message  string
		text     string
		expected string
	}{
		"remember simple": {
			remember: "/remember !calltext text",
			callout:  "calltext",
			message:  "!calltext foo",
			text:     "text",
			expected: "text",
		},
		"remember emoji": {
			remember: "/remember !callmoji text 🐶🎉🥳",
			callout:  "callmoji",
			message:  "!callmoji foo",
			text:     "text 🐶🎉🥳",
			expected: "text 🐶🎉🥳",
		},
		"remember text with symbols": {
			remember: "/remember !marrano %, you are very marrano!",
			callout:  "marrano",
			message:  "!marrano foo",
			text:     "%, you are very marrano!",
			expected: "foo, you are very marrano!",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			res, err := tryHandle(t, `{
				"message": {
					"chat": {"id":"`+gid+`"},
					"text": "`+test.remember+`"
				}
			}`)
			assert.NilError(t, err)
			assert.Check(t, *res.Text == textRemember)

			c := &db.Callout{
				GID:     gid,
				Callout: test.callout,
			}
			err = db.SelectOneCallout(context.TODO(), c)
			assert.NilError(t, err)
			assert.Equal(t, c.Text, test.text)

			res, err = tryHandle(t, `{
				"message": {
					"chat": { "id": "`+gid+`" },
					"text": "`+test.message+`"
				}
			}`)
			assert.NilError(t, err)
			assert.Check(t, *res.Text == test.expected)
		})
	}
}

func TestHandlerAbraxas(t *testing.T) {
	gid := "123456"

	prepareDB(t, gid)
	defer db.Close()

	testsRemember := map[string]struct {
		message string
		kind    string
		trigger string
	}{
		"video": {
			message: "/remember clobber video",
			kind:    "video",
			trigger: "clobber",
		},
		"animation": {
			message: "/remember animation animation",
			kind:    "video",
			trigger: "animation",
		},
		"photo": {
			message: "/remember photo photo",
			trigger: "photo",
			kind:    "photo",
		},
		"default": {
			message: "/remember moo asdfg",
			trigger: "moo",
			kind:    "photo",
		},
	}

	for name, test := range testsRemember {
		test := test
		t.Run(name, func(t *testing.T) {
			res, err := tryHandle(t, `{
				"message": {
					"chat": {"id":"`+gid+`"},
					"text": "`+test.message+`"
				}
			}`)
			assert.NilError(t, err)
			assert.Check(t, *res.Text == textRemember)

			a := &db.Abraxas{GID: gid, Abraxas: test.trigger}
			err = db.SelectOneAbraxasByAbraxas(context.TODO(), a)
			assert.NilError(t, err)
			assert.Equal(t, a.Kind, test.kind)
		})
	}

	dummyMedia := []*db.Media{
		{GID: gid, Data: "123456", Kind: "photo", Description: ""},
		{GID: gid, Data: "234567", Kind: "photo", Description: ""},
		{GID: gid, Data: "345678", Kind: "video", Description: ""},
		{GID: gid, Data: "456789", Kind: "video", Description: ""},
	}
	for _, m := range dummyMedia {
		err := db.InsertMedia(context.TODO(), m)
		assert.NilError(t, err)
	}
	m := &db.Media{
		GID:  gid,
		Data: "123456",
		Kind: "photo",
	}
	err := db.SelectOneMediaByData(context.TODO(), m)
	assert.NilError(t, err)
	assert.Check(t, m.Data == "123456")

	testsInvoke := map[string]struct {
		message string
		trigger string
		method  string
	}{
		"nothing": {
			message: "some random people chatting harmlessly",
			trigger: "some",
			method:  "",
		},
		"photo simple": {
			message: "photo",
			trigger: "clobber",
			method:  "sendPhoto",
		},
		"photo complex": {
			message: "photo thing",
			trigger: "photo",
			method:  "sendPhoto",
		},
		"video simple": {
			message: "clobber",
			trigger: "clobber",
			method:  "sendVideo",
		},
		"video complex": {
			message: "clobber thing",
			trigger: "clobber",
			method:  "sendVideo",
		},
	}
	for name, test := range testsInvoke {
		test := test
		t.Run(name, func(t *testing.T) {
			res, err := tryHandle(t, `{
				"message": {
					"chat": {"id":"`+gid+`"},
					"text": "`+test.message+`"
				}
			}`)
			assert.NilError(t, err)
			assert.Check(t, res != nil)
			t.Log("message", test.message, "method", res.Method)
			assert.Check(t, res.Method == test.method)
		})
	}
}