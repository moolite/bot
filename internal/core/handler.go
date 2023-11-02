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
	CmdRemember = iota + 1
	CmdForget
	CmdLink
	CmdNote
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

	if !p.Exists("message", "chat", "id") {
		return nil, ErrParseNoChatID
	}
	chatId := string(p.GetStringBytes("message", "chat", "id"))

	var text string
	if p.Exists("message", "text") {
		text = string(p.GetStringBytes("message", "text"))
	} else if p.Exists("message", "caption") {
		text = string(p.GetStringBytes("message", "caption"))
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
		switch inc.Command {
		case CmdRemember:
			if p.Exists("message", "text") {

			} else { // Most likely a forwarded message or media
				media := &db.Media{
					GID: chatId,
				}

				if p.Exists("message", "photo") {
					media.Kind = "photo"
					media.Data = string(p.GetStringBytes("message", "photo", "0", "file_id"))

				} else if p.Exists("message", "animation") {
					media.Kind = "animation"
					p.GetStringBytes("message", "animation", "0", "file_id")

				} else if p.Exists("message", "video") {
					media.Kind = "video"
					p.GetStringBytes("message", "video", "0", "file_id")
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
			if p.Exists("message", "photo") {
				media := &db.Media{
					GID: chatId,
				}
				sizes := p.GetArray("message", "photo")
				if len(sizes) > 0 {
					media.Data = string(sizes[0].GetStringBytes("file_id"))
				}

				if err := db.Query(ctx, dbc, media.Delete()); err != nil {
					log.Error().Err(err).Msg("error deleting media")
				}
			}

			res.SendMessage().
				SetText(textForget)

		case CmdLink:
			res.SendMessage().
				SetText(textRemember)

		case CmdNote:
			res.SendMessage().
				SetText(textForget)

		case CmdDice:
			res.SendDice()

		default:
			res.SendMessage().
				SetText(text404)
		}
	}

	res.SetChatID(string(chatId))

	return res.Marshal()
}
