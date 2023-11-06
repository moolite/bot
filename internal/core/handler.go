package core

import (
	"context"
	"database/sql"
	"errors"
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
	gid := string(p.GetStringBytes("chat", "id"))

	var text string
	if p.Exists("text") {
		text = string(p.GetStringBytes("text"))
	} else if p.Exists("caption") {
		text = string(p.GetStringBytes("caption"))
	}

	inc := parseText(text)

	switch inc.Kind {
	case KindTrigger:
		trigger := &db.Abraxas{GID: gid, Abraxas: inc.Abraxas}

		// FIXME what if there is no trigger?
		if err := db.SelectOneAbraxas(ctx, trigger); err != nil {
			log.Error().Err(err).Msg("abraxoides get not found")
		}

		if trigger.Kind == "" {
			log.Debug().Interface("trigger", trigger).Msg("trigger has no kind")
			return nil, nil
		}

		media := &db.Media{
			GID:  gid,
			Kind: trigger.Kind,
		}

		if err := db.SelectRandomMedia(ctx, media); err != nil {
			log.Error().Err(err).Msg("error selecting random media")
		}

		switch media.Kind {
		case "video":
			res.SendVideo(media.Data)
		case "animation":
			res.SendAnimation(media.Data)
		case "photo":
			res.SendPhoto(media.Data)
		default:
			return nil, nil
		}

	case KindCallout:
		callout := &db.Callout{
			GID:     gid,
			Callout: inc.Abraxas,
		}

		if err := db.SelectOneCallout(ctx, callout); err != nil {
			log.Error().Err(err).Msg("callout query error")
			return nil, err
		} else if callout.Text != "" {
			res.SendMessage().
				SetText(strings.Replace(callout.Text, "%", inc.Rest, -1))
		} else {
			return nil, nil
		}

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
				var kind string
				var data string

				if p.Exists("photo") { // Photo
					kind = "photo"
					data = string(p.GetStringBytes("photo", "0", "file_id"))

				} else if p.Exists("message", "animation") { // Animations
					kind = "animation"
					data = string(p.GetStringBytes("animation", "0", "file_id"))

				} else if p.Exists("message", "video") { // Video
					kind = "video"
					data = string(p.GetStringBytes("video", "0", "file_id"))

				}

				m := &db.Media{
					GID:         gid,
					Kind:        kind,
					Data:        data,
					Description: inc.Rest,
				}

				res.SendMessage()

				if err := db.InsertMedia(ctx, m); err != nil {
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
					GID: gid,
				}

				if p.Exists("photo", "0", "file_id") {
					media.Data = string(p.GetStringBytes("photo", "0", "file_id"))

				} else if p.Exists("animation", "0", "file_id") {
					media.Data = string(p.GetStringBytes("animation", "0", "file_id"))

				} else if p.Exists("video", "0", "file_id") {
					media.Data = string(p.GetStringBytes("video", "0", "file_id"))
				}

				if err := db.DeleteMedia(ctx, media); err != nil {
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

			switch inc.Operation {
			case OpAdd:
				entities := p.GetArray("message", "entities")
				var url string
				for _, ent := range entities {
					if ent.Exists("url") {
						url = string(ent.GetStringBytes("url"))
					}
				}

				link := &db.Link{
					GID:  gid,
					URL:  url,
					Text: text,
				}

				if err := db.InsertLink(ctx, link); err != nil {
					log.Error().Err(err).Msg("error inserting link")

					res.SetText(textErr)
				} else {
					res.SetText(textRemember)
				}

			case OpRem:
				entities := p.GetArray("message", "entities")
				var url string
				for _, ent := range entities {
					if ent.Exists("url") {
						url = string(ent.GetStringBytes("url"))
					}
				}

				link := &db.Link{
					GID: gid,
					URL: url,
				}

				if err := db.DeleteLink(ctx, link); err != nil {
					log.Error().Err(err).Msg("error inserting link")

					res.SetText(textErr)
				} else {
					res.SetText(textForget)
				}

			default:
				if links, err := db.SearchLinks(ctx, gid, text); err != nil {
					log.Error().Err(err).Msg("error searching links")

					res.SetText(textErr)

				} else if len(links) == 0 {
					res.SetText(text404)

				} else {
					var buttons []telegram.URLButton
					for _, v := range links {
						buttons = append(buttons, telegram.URLButton{
							URL:  v.URL,
							Text: v.Text,
						})
					}
					res.SetText("links:").
						SetLinks(buttons)
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

	res.SetChatID(string(gid))

	return res.Marshal()
}
