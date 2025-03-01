package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/matryer/is"
)

func TestRunHandlers(t *testing.T) {
	is := is.New(t)
	b := &Bot{onMessageHandler: noopHandler}

	testCases := map[string]struct {
		Type   UpdateHandlerType
		Param  string
		Update *Update
	}{
		"startsWith": {
			Type:   UPD_STARTSWITH,
			Param:  "/foo",
			Update: &Update{Message: &Message{Text: "/foo bar baz"}},
		},
		"contains": {
			Type:   UPD_CONTAINS,
			Param:  "foo",
			Update: &Update{Message: &Message{Text: "maybe foo bar baz"}},
		},
		"regexp": {
			Type:   UPD_REGEXP,
			Param:  "fo+\\sbar",
			Update: &Update{Message: &Message{Text: "moo fooooo bar baz"}},
		},
		"media photo": {
			Type:   UPD_MEDIAPHOTO,
			Update: &Update{Message: &Message{Photo: []*PhotoSize{{FileID: "132456"}}}},
		},
		"media video": {
			Type:   UPD_MEDIAVIDEO,
			Update: &Update{Message: &Message{Video: &Video{FileID: "123456"}}},
		},
		"media document": {
			Type:   UPD_MEDIADOCUMENT,
			Update: &Update{Message: &Message{Document: &Document{FileID: "123456"}}},
		},
		"callback": {
			Type:   UPD_CALLBACK,
			Update: &Update{CallbackQuery: &CallbackQuery{ID: "foo"}},
		},
		"media wildcard": {
			Type:   UPD_WILDCARD,
			Update: &Update{Message: &Message{Text: "any"}},
		},
	}

	for name, testcase := range testCases {
		t.Run(name, func(t *testing.T) {
			is := is.New(t)
			resp := &Sendable{
				ChatID: 123456,
				Text:   "foob!",
			}
			b.handlers = []*UpdateHandler{{
				Type:  testcase.Type,
				Param: testcase.Param,
				Fn: func(ctx context.Context, bot *Bot, update *Update) (*Sendable, error) {
					return resp, nil
				},
				rx: regexp.MustCompile(testcase.Param),
			}}

			s, err := b.RunHandlers(context.TODO(), testcase.Update)
			is.NoErr(err)
			is.True(s != nil)
			is.True(s.ChatID == resp.ChatID)
			is.True(s.Text == resp.Text)
		})
	}

	t.Run("stack", func(t *testing.T) {
		is := is.New(t)

		testCases = map[string]struct {
			Type   UpdateHandlerType
			Param  string
			Update *Update
		}{
			"fake 01": {Type: UPD_STARTSWITH, Param: "/foo", Update: &Update{Message: &Message{Text: ""}}},
			"fake 02": {Type: UPD_STARTSWITH, Param: "/foo", Update: &Update{Message: &Message{Text: ""}}},
			"fake 03": {Type: UPD_STARTSWITH, Param: "/foo", Update: &Update{Message: &Message{Text: ""}}},
		}
		b.handlers = make([]*UpdateHandler, 0)
		for _, testcase := range testCases {
			b.RegisterHandlers(&UpdateHandler{
				Type:  testcase.Type,
				Param: testcase.Param,
				Fn: func(ctx context.Context, bot *Bot, update *Update) (*Sendable, error) {
					return nil, nil
				},
				rx: regexp.MustCompile(testcase.Param),
			})
		}

		resp := &Sendable{Method: "sendMessage", Text: "this is a test"}
		b.handlers = append([]*UpdateHandler{{
			Type:  UPD_STARTSWITH,
			Param: "/foo",
			Fn: func(ctx context.Context, bot *Bot, update *Update) (*Sendable, error) {
				return resp, nil
			},
		}}, b.handlers...)

		x, err := b.RunHandlers(context.TODO(), &Update{Message: &Message{Text: "/foo"}})
		is.NoErr(err)
		is.True(x != nil)
		is.True(x.Text == resp.Text)
	})
}

func parseSendable(body []byte) *Sendable {
	res := &Sendable{}
	if err := json.Unmarshal(body, res); err != nil {
		panic(err)
	}
	return res
}

func mustJson(x any) []byte {
	r, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return r
}

func TestHttpHandler(t *testing.T) {
	is := is.New(t)
	is.True(true)

	b := &Bot{onMessageHandler: noopHandler}
	httpHandler := b.HttpHandler(slog.Default())

	testCases := map[string]struct {
		Update *Update
		Method string
	}{
		"simple message": {
			Update: &Update{UpdateID: 123456, Message: &Message{Text: "some message"}},
			Method: "",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			is := is.New(t)
			body := bytes.NewReader(mustJson(testCase.Update))
			res := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/t/bot", body)
			httpHandler(res, req)

			is.True(res.Code == http.StatusOK)

			s := parseSendable(res.Body.Bytes())
			is.True(s.Method == testCase.Method)
		})
	}
}
