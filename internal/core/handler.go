package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/dicer"
	"github.com/moolite/bot/internal/telegram"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fastjson"
)

var (
	ErrParseNoChatID  error = errors.New("chat id not defined")
	ErrParseNoMessage error = errors.New("message not defined")
)

var (
	re = regexp.MustCompile(
		`(([/!])?(?P<abraxas>[^@\s]+)(@[^\s]+)?)\s*(?P<rest>(?P<add>(\+|add|ricorda|record|put))?(?P<rem>(\-|rem|del|dimentica|forget|drop))?.*)`)

	reCallout = regexp.MustCompile(
		`(!)(?P<abraxas>[^\s]+)\s+(?P<rest>.+)`)

	reBackup = regexp.MustCompile(`^(back(up)?)$`)
	reMember = regexp.MustCompile(`^(\+|add|remember|ricorda)[^\s]*$`)
	reForget = regexp.MustCompile(`^(\-|del|forget|dimentica)[^\s]*$`)
	reDice   = regexp.MustCompile(`^(d|dice|r|roll)[^\s]*$`)
	rePhoto  = regexp.MustCompile(`^(p|photo|f|foto)[^\s]*$`)
	reVideo  = regexp.MustCompile(`^(v|video|clip|ani)[^\s]*$`)
)

var (
	textRemember string = "ho imparato _qualcosa_."
	textForget   string = "ho dimenticato _qualcosa_."
	textErr      string = "qualcosa è andato storto."
	text404      string = "non ho capito."
)

const (
	_ = iota
	CmdBackup
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

				if reBackup.MatchString(r.Abraxas) {
					r.Command = CmdBackup
				} else if reMember.MatchString(r.Abraxas) {
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

	if strings.HasPrefix(r.Rest, "!") {
		for _, m := range reCallout.FindAllStringSubmatch(r.Rest, -1) {
			for i, name := range reCallout.SubexpNames() {
				switch name {
				case "abraxas":
					r.Abraxas = m[i]
				case "rest":
					r.Rest = m[i]
				}
			}
		}
		r.Kind = KindCallout
	}

	return r
}

func Handler(p *fastjson.Value) (*telegram.WebhookResponse, error) {
	ctx := context.Background()
	res := new(telegram.WebhookResponse)

	// Early exit
	if !p.Exists("message") {
		return res, ErrParseNoMessage
	}

	// move parser into the "p" part
	p = p.Get("message")

	if !p.Exists("chat", "id") {
		return res, ErrParseNoChatID
	}
	gid := string(p.GetStringBytes("chat", "id"))
	res.SetChatID(string(gid))

	var text string
	if p.Exists("text") {
		text = string(p.GetStringBytes("text"))
	} else if p.Exists("caption") {
		text = string(p.GetStringBytes("caption"))
	}

	// Use forwarded messages as base for data
	if p.Exists("message", "reply_to_message") {
		p = p.Get("message", "reply_to_message")
	}

	inc := parseText(text)

	switch inc.Kind {
	case KindTrigger:
		trigger := &db.Abraxas{
			GID:     gid,
			Abraxas: inc.Abraxas,
		}

		// FIXME what if there is no trigger?
		if err := db.SelectOneAbraxas(ctx, trigger); errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Msg("abraxoides get not found")
			return res, nil
		} else if err != nil {
			return res, err
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
		switch inc.Command {
		case CmdRemember:
			if p.Exists("photo") || p.Exists("animation") || p.Exists("video") {
				m := &db.Media{
					GID:         gid,
					Description: inc.Rest,
				}

				if p.Exists("photo") { // Photo
					m.Kind = "photo"
					m.Data = string(p.GetStringBytes("photo", "0", "file_id"))

				} else if p.Exists("animation") { // Animations
					m.Kind = "animation"
					m.Data = string(p.GetStringBytes("animation", "0", "file_id"))

				} else if p.Exists("video") { // Video
					m.Kind = "video"
					m.Data = string(p.GetStringBytes("video", "0", "file_id"))
				} else {
					return res, fmt.Errorf("something is wrong")
				}

				log.Error().Str("kind", m.Kind).Str("data", m.Data).Msg("new media")

				res.SendMessage()

				err := db.InsertMedia(ctx, m)
				if err != nil {
					log.Error().Err(err).Msg("error inserting new media")
					return res.SetText(textErr), err
				}

				return res.SetText(textRemember), nil

			} else if inc.Kind == KindCallout && len(inc.Abraxas) > 0 {
				c := &db.Callout{GID: gid, Callout: inc.Abraxas, Text: inc.Rest}

				err := db.InsertCallout(ctx, c)
				if err != nil {
					log.Error().
						Err(err).
						Str("callout", inc.Abraxas).
						Str("rest", inc.Rest).
						Msg("error inserting callout.")
					return res.SetText(textErr), nil
				}
			}

			log.Error().
				Bool("photo", p.Exists("photo")).
				Bool("video", p.Exists("video")).
				Bool("animation", p.Exists("animation")).
				Str("p", string(p.GetStringBytes("video", "0", "file_id"))).
				Msg("WRTFD")

			return res.SendMessage().SetText(textErr), nil

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

	return res.Empty(), nil
}
