package core

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/matryer/is"
	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/pkg/tg"
)

type testCase struct {
	upd    *tg.Update
	method string
	text   string
	isNil  bool
	isErr  bool
}

func pprint(t *testing.T, s interface{}) {
	t.Helper()

	m, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		t.Log(err)
	}
	t.Log(string(m))
}

func newBot() *tg.Bot {
	bot, err := tg.New("123456", "https://bot.marrani.lol")
	if err != nil {
		panic(err)
	}
	return bot
}

func updMessage(text string) *tg.Update {
	return &tg.Update{
		UpdateID: -123456,
		Message: &tg.Message{
			ID:   1,
			Chat: tg.Chat{ID: 123456},
			Text: text,
		},
	}
}

func updMediaVideo(text string, id string) *tg.Update {
	u := updMessage(text)
	u.Message.Video = &tg.Video{FileID: id}
	return u
}

func updMediaPhoto(text string, id string) *tg.Update {
	u := updMessage(text)
	u.Message.Photo = []*tg.PhotoSize{{FileID: id}}
	return u
}

func TestBotHandlers(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	// t.Parallel()
	is := is.New(t)

	is.NoErr(db.Open(":memory:"))
	is.NoErr(db.Migrate())
	defer db.Close()

	testCases := map[string]*testCase{
		"remember":         {upd: updMessage("/remember"), method: "sendMessage", isNil: true},
		"remember video":   {upd: updMediaVideo("/remember video", "video123"), method: "setMessageReaction"},
		"remember photo":   {upd: updMediaPhoto("/remember photo", "photo123"), method: "setMessageReaction"},
		"remember symbols": {upd: updMediaPhoto("/remember a b% %c?!", "photo234"), method: "setMessageReaction"},
		"ricorda alias":    {upd: updMediaVideo("/ricorda alias", "video546"), method: "setMessageReaction"},
		"add alias":        {upd: updMediaVideo("/add alias", "video546"), method: "setMessageReaction"},
		"touch alias":      {upd: updMediaVideo("/touch alias", "video546"), method: "setMessageReaction"},
	}
	for text, tc := range testCases {
		t.Run(text, func(t *testing.T) {
			is := is.New(t)
			ctx := context.TODO()
			b := newBot()
			s, err := MediaRememberCommand(ctx, b, tc.upd)

			is.NoErr(err)
			if tc.isNil {
				is.True(s == nil)
			} else {
				is.True(s != nil)
				is.True(s.Method == tc.method)
			}
		})
	}

	testCases = map[string]*testCase{
		"dice":       {upd: updMessage("/dice"), isNil: true},
		"dice alias": {upd: updMessage("/d 6d8"), method: "sendMessage"},
		"dice wrong": {upd: updMessage("/d wrong"), isNil: true},
	}
	for text, tc := range testCases {
		t.Run(text, func(t *testing.T) {
			is := is.New(t)
			ctx := context.TODO()
			b := newBot()
			s, err := DiceCommand(ctx, b, tc.upd)
			is.NoErr(err)
			if tc.isNil {
				is.True(s == nil)
			} else {
				is.True(s != nil)
				is.True(s.Method == tc.method)
			}
			pprint(t, s)
		})
	}

	testCases = map[string]*testCase{
		"callout ok":  {upd: updMessage("/callout marrano %, sei un marrano!"), method: "setMessageReaction"},
		"callout ok2": {upd: updMessage("/callout paris %, sei helpy come haris pilton!"), method: "setMessageReaction"},
		"callout err": {upd: updMessage("/callout %, sei un marrano!"), isNil: true},
	}
	for text, tc := range testCases {
		t.Run(text, func(t *testing.T) {
			is := is.New(t)
			ctx := context.TODO()
			b := newBot()
			s, err := CalloutCommand(ctx, b, tc.upd)
			is.NoErr(err)
			if tc.isNil {
				is.True(s == nil)
			} else {
				is.True(s != nil)
				is.True(s.Method == tc.method)
			}
			pprint(t, s)
		})
	}

	testCases = map[string]*testCase{
		"callout msg ok":        {upd: updMessage("!marrano bot"), method: "sendMessage", text: "\u003cb\u003ebot\u003c/b\u003e, sei un marrano!"},
		"callout msg ok2":       {upd: updMessage("!paris bot"), method: "sendMessage", text: "\u003cb\u003ebot\u003c/b\u003e, sei helpy come haris pilton!"},
		"callout msg missing":   {upd: updMessage("!nonexistant callout"), isNil: true, isErr: true},
		"callout msg wrong":     {upd: updMessage("!asd callout"), isNil: true, isErr: true},
		"callout msg nocallout": {upd: updMessage("!paris"), isNil: true, isErr: true},
	}
	for text, tc := range testCases {
		t.Run(text, func(t *testing.T) {
			is := is.New(t)
			ctx := context.TODO()
			b := newBot()
			s, err := CalloutMessage(ctx, b, tc.upd)
			if !tc.isErr {
				is.NoErr(err)
			}

			pprint(t, s)
			if tc.isNil {
				is.True(s == nil)
			} else {
				is.True(s != nil)
				is.True(s.Method == tc.method)
			}
			if tc.text != "" {
				is.True(s.Text == tc.text)
			}
		})
	}
}

// "{\"chat_id\":-284819895,
//   \"parse_mode\":\"html\",
//   \"reply_markup\":\"{\\\"inline_keyboard\\\":[[{\\\"text\\\":\\\"❤\\\",\\\"callback_data\\\":\\\"media:up\\\"},{\\\"text\\\":\\\"\\\\u003cb\\\\u003e\\\\u003ci\\\\u003e0\\\\u003c/i\\\\u003e\\\\u003c/b\\\\u003e\\\"},{\\\"text\\\":\\\"💔\\\",\\\"callback_data\\\":\\\"media:down\\\"}]]}\",
//   \"caption\":\"Killer doccia\",
//   \"photo\":\"AgACAgQAAxkBAAEEd4JhAm4LVrZRI5keQyF3OMub96Zk7gAC-bUxGznCGVDc7b7LG-UH6gEAAwIAA3MAAyAE\"
// }"
