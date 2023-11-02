package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/dicer"
	"github.com/moolite/bot/internal/telegram"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fastjson"
)

var (
	ErrParseNoChatID error = errors.New("chat id not defined")
)

var (
	re = regexp.MustCompile(
		`(([/!])?(?P<abraxas>[^@\s]+)(@[^\s]+)?)\s*(?P<rest>(?P<add>(\+|add|ricorda|record|put))?(?P<rem>(\-|rem|del|dimentica|forget|drop))?.*)`)

	reMember = regexp.MustCompile(`^(\+|add|ricorda)$`)
	reForget = regexp.MustCompile(`^(\-|del|dimentica)$`)
	reDice   = regexp.MustCompile(`^(d|dice|r|roll)$`)
	rePhoto  = regexp.MustCompile(`^(p|photo|f|foto)$`)
	reVideo  = regexp.MustCompile(`^(v|video|clip|ani)$`)
)

var (
	textRemember string = "ho imparato _qualcosa_."
	textForget   string = "ho dimenticato _qualcosa_."
	textErr      string = "qualcosa è andato storto."
	text404      string = "non ho capito."
)

const (
	_ = iota
	CmdRemember
	CmdForget
	CmdLink
	CmdDice
	CmdVideo
	CmdPhoto
	_
	OpAdd
	OpRem
	_
	KindCommand
	KindCallout
	KindTrigger
)

type BotRequest struct {
	Kind      int
	Operation int
	Command   int
	Abraxas   string
	Rest      string
}

func isCallout(text string) bool {
	return strings.HasPrefix(text, "!")
}

func parseText(text string) *BotRequest {
	r := &BotRequest{}

	if strings.HasPrefix(text, "/") {
		r.Kind = KindCommand
	} else if strings.HasPrefix(text, "!") {
		r.Kind = KindCallout
	} else {
		r.Kind = KindTrigger
	}

	for _, m := range re.FindAllStringSubmatch(text, -1) {
		for i, name := range re.SubexpNames() {

			fmt.Fprintf(os.Stderr, "parsed : name:%s m:'%s'\n", name, m[i])

			switch name {
			case "abraxas":
				r.Abraxas = m[i]

				if reMember.MatchString(r.Abraxas) {
					r.Command = CmdRemember
				} else if reForget.MatchString(r.Abraxas) {
					r.Command = CmdForget
				} else if reDice.MatchString(r.Abraxas) {
					r.Command = CmdDice
				} else if rePhoto.MatchString(r.Abraxas) {
					r.Command = CmdPhoto
				} else if reVideo.MatchString(r.Abraxas) {
					r.Command = CmdVideo
				}

			case "rest":
				r.Rest = m[i]

			case "add":
				if m[i] != "" {
					r.Operation = OpAdd
				}

			case "rem":
				if m[i] != "" {
					r.Operation = OpRem
				}
			}
		}
	}

	return r
}

func Handler(p *fastjson.Value, dbc *sql.DB) ([]byte, error) {
	ctx := context.Background()

	res := new(telegram.WebhookResponse)

	// Early exit
	if !p.Exists("message") {
		return res.
			SetText(textErr).
			SendMessage().
			Marshal()
	}

	// move parser into the "message" part
	p = p.Get("message")

	if !p.Exists("chat", "id") {
		return nil, ErrParseNoChatID
	}
	chatId := string(p.GetStringBytes("chat", "id"))

	var text string
	if p.Exists("text") {
		text = string(p.GetStringBytes("text"))
	} else if p.Exists("caption") {
		text = string(p.GetStringBytes("caption"))
	}

	inc := parseText(text)

	switch inc.Kind {
	case KindTrigger:
		t := &db.Abraxoides{
			Abraxas: inc.Abraxas,
		}

		// FIXME what if there is no trigger?
		if err := db.QueryOne[db.Abraxoides](ctx, dbc, t.One(), t); err != nil {
			log.Error().Err(err).Msg("abraxoides get not found")
		}

	case KindCallout:
		callout := &db.Callout{
			Callout: inc.Abraxas,
		}

		if err := db.QueryOne[db.Callout](ctx, dbc, callout.One(), callout); err != nil {
			log.Error().Err(err).Msg("callout query error")
			return nil, err
		}

		res.SendMessage().
			SetText(strings.Replace(callout.Text, "%", inc.Rest, -1))

	case KindCommand:
		// Use forwarded messages as base for data
		if p.Exists("message", "reply_to_message") {
			p = p.Get("message", "reply_to_message")
		} else {
			p = p.Get("message")
		}

		switch inc.Command {
		case CmdRemember:
			if p.Exists("text") {

			} else { // Most likely media
				media := &db.Media{
					GID: chatId,
				}

				if p.Exists("photo") { // Photo
					media.Kind = "photo"
					media.Data = string(p.GetStringBytes("photo", "0", "file_id"))

				} else if p.Exists("message", "animation") { // Animations
					media.Kind = "animation"
					media.Data = string(p.GetStringBytes("animation", "0", "file_id"))

				} else if p.Exists("message", "video") { // Video
					media.Kind = "video"
					media.Data = string(p.GetStringBytes("video", "0", "file_id"))

				}

				media.Description = inc.Rest

				res.SendMessage()

				if err := db.Query(ctx, dbc, media.Insert()); err != nil {
					log.Error().Err(err).Msg("error inserting new media")
					res.SetText(textErr)
				} else {
					res.SetText(textRemember)
				}
			}

		case CmdForget:
			if p.Exists("message", "text") {
			} else { // Likely media
				media := &db.Media{
					GID: chatId,
				}

				if p.Exists("photo", "0", "file_id") {
					media.Data = string(p.GetStringBytes("photo", "0", "file_id"))

				} else if p.Exists("animation", "0", "file_id") {
					media.Data = string(p.GetStringBytes("animation", "0", "file_id"))

				} else if p.Exists("video", "0", "file_id") {
					media.Data = string(p.GetStringBytes("video", "0", "file_id"))
				}

				if err := db.Query(ctx, dbc, media.Delete()); err != nil {
					log.Error().Err(err).Msg("error deleting media")

					res.SetText(textErr)
				} else {
					res.SetText(textForget)
				}
			}

			res.SendMessage().
				SetText(textForget)
		case CmdLink:
			res.SendMessage()

			l := &db.Links{
				GID:  chatId,
				Text: text,
			}

			switch inc.Operation {
			case OpAdd:

				entities := p.GetArray("message", "entities")
				for _, ent := range entities {
					if ent.Exists("url") {
						l.URL = string(ent.GetStringBytes("url"))
					}
				}

				if err := db.Query(ctx, dbc, l.Insert()); err != nil {
					log.Error().Err(err).Msg("error inserting link")

					res.SetText(textErr)
				} else {
					res.SetText(textRemember)
				}

			case OpRem:
				entities := p.GetArray("message", "entities")
				for _, ent := range entities {
					if ent.Exists("url") {
						l.URL = string(ent.GetStringBytes("url"))
					}
				}

				if err := db.Query(ctx, dbc, l.Insert()); err != nil {
					log.Error().Err(err).Msg("error inserting link")

					res.SetText(textErr)
				} else {
					res.SetText(textForget)
				}

			default:
				l := &db.Links{}
				var searchResults []*db.Links

				if err := db.QueryMany[db.Links](ctx, dbc, l.Search(), searchResults); err != nil {
					log.Error().Err(err).Msg("error searching links")
					res.SetText(textErr)

				} else if len(searchResults) == 0 {
					res.SetText(text404)

				} else {
					var buttons []telegram.URLButton
					for _, v := range searchResults {
						buttons = append(buttons, telegram.URLButton{
							URL:  v.URL,
							Text: v.Text,
						})
					}
					res.SetText("links:").SetLinks(buttons)
				}
			}

		case CmdDice:
			dies := dicer.New(inc.Rest)
			t := ""
			for _, die := range dies {
				t += die.Markdown()
			}

			res.SendDice().
				SetText(t)

		default:
			res.SendMessage().
				SetText(text404)
		}
	}

	res.SetChatID(string(chatId))

	return res.Marshal()
}
