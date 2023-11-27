package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/dicer"
	"github.com/moolite/bot/internal/statistics"
	"github.com/moolite/bot/internal/telegram"
	"github.com/valyala/fastjson"
)

var (
	ErrParseNoChatID  error = errors.New("chat id not defined")
	ErrParseNoMessage error = errors.New("message not defined")
)

var (
	re = regexp.MustCompile(
		`^(([/!])?(?P<abraxas>[^@\s]+)(@[^\s]+)?)\s*(?P<rest>(?P<add>(\+|add|ricorda|record|put))?(?P<rem>(\-|rem|del|dimentica|forget|drop))?.*)\s*$`)

	reMemberCmd = regexp.MustCompile(
		`^!?(?P<abraxas>[^\s]+)\s*(?P<rest>.+)?\s*$`)

	reBackup = regexp.MustCompile(`^(back(up)?)$`)
	reMember = regexp.MustCompile(`^(\+|add|remember|ricorda)[^\s]*$`)
	reForget = regexp.MustCompile(`^(\-|del|forget|dimentica)[^\s]*$`)
	reDice   = regexp.MustCompile(`^(d|dice|r|roll)[^\s]*$`)
	rePhoto  = regexp.MustCompile(`^(p|photo|f|foto)[^\s]*$`)
	reVideo  = regexp.MustCompile(`^(v|video|clip|ani)[^\s]*$`)
	reLink   = regexp.MustCompile(`^(l|link|url)[^\s]*$`)
)

var (
	textRemember string = "ho imparato _qualcosa_."
	textForget   string = "ho dimenticato _qualcosa_."
	textErr      string = "qualcosa è andato storto."
	text404      string = "non ho capito."
)

var (
	ErrNotFound    error = errors.New("not found")
	ErrURLNotFound error = errors.New("url not found")
)

const (
	_None = iota
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
	KindCalloutCmd
	KindTrigger
	KindTriggerCmd
)

type BotRequest struct {
	Kind      int
	Operation int
	Command   int
	Abraxas   string
	Rest      string
	// Remember
	RememberKind    int
	RememberAbraxas string
	RememberRest    string
}

func isCallout(text string) bool {
	return strings.HasPrefix(text, "!")
}

func getURL(p *fastjson.Value) (string, error) {
	entities := p.GetArray("entities")
	var url string
	for _, ent := range entities {
		if ent.Exists("url") {
			url = string(ent.GetStringBytes("url"))
		}
	}
	if url == "" {
		return url, ErrURLNotFound
	}
	return url, nil
}

func parseText(text string) *BotRequest {
	r := &BotRequest{
		Command:   _None,
		Kind:      _None,
		Operation: _None,
	}

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
				} else if reLink.MatchString(r.Abraxas) {
					r.Command = CmdLink
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

	if r.Command == CmdRemember {
		if strings.HasPrefix(r.Rest, "!") {
			r.RememberKind = KindCalloutCmd
		} else {
			r.RememberKind = KindTriggerCmd
		}
		hasRest := false
		for _, m := range reMemberCmd.FindAllStringSubmatch(r.Rest, -1) {
			for i, name := range reMemberCmd.SubexpNames() {
				switch name {
				case "abraxas":
					r.RememberAbraxas = m[i]
				case "rest":
					hasRest = true
					r.RememberRest = m[i]
				}
			}
		}

		// cleanup rest when omitted
		if !hasRest {
			r.RememberRest = ""
		}
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

	username := "dummy"
	if p.Exists("from", "username") {
		username = string(p.GetStringBytes("from", "username"))
	}

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

	go statistics.ApplyTriggers(gid, username, text)

	inc := parseText(text)

	switch inc.Kind {
	case KindTrigger:
		trigger := &db.Abraxas{
			GID:     gid,
			Abraxas: inc.Abraxas,
		}

		err := db.SelectOneAbraxasByAbraxas(ctx, trigger)
		if errors.Is(err, sql.ErrNoRows) {
			return res, nil
		} else if err != nil {
			slog.Error("handler SelectOneAbraxasByAbraxas error", "abraxas", inc.Abraxas, "err", err)
			return res, err
		} else if trigger.Kind == "" {
			slog.Debug("handler trigger has no kind", "trigger", trigger)
			return nil, nil
		}

		slog.Debug("trigger", "kind", trigger.Kind)

		media := &db.Media{
			GID:  gid,
			Kind: trigger.Kind,
		}

		err = db.SelectRandomMedia(ctx, media)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return res.SendMessage().
					SetText(text404), err
			} else {
				slog.Error("error selecting random media", "err", err)
				return res.Empty(), err
			}
		}

		switch media.Kind {
		case "video":
			res.SendVideo(media.Data)
		case "photo":
			res.SendPhoto(media.Data)
		}

		return res, nil

	case KindCallout:
		callout := &db.Callout{
			GID:     gid,
			Callout: inc.Abraxas,
		}

		if err := db.SelectOneCallout(ctx, callout); err != nil {
			slog.Error("callout query error", "err", err)

			if errors.Is(err, sql.ErrNoRows) {
				return res.SendMessage().SetText(textErr), nil
			}
			return nil, err
		} else if callout.Text != "" {
			return res.SendMessage().
				SetText(strings.Replace(callout.Text, "%", inc.Rest, -1)), nil
		}

		return res, nil

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

				} else if p.Exists("animation") { // Animations are saved as video
					m.Kind = "video"
					m.Data = string(p.GetStringBytes("animation", "0", "file_id"))

				} else if p.Exists("video") { // Video
					m.Kind = "video"
					m.Data = string(p.GetStringBytes("video", "0", "file_id"))
				} else {
					return res, fmt.Errorf("something is wrong")
				}

				res.SendMessage()

				err := db.InsertMedia(ctx, m)
				if err != nil {
					slog.Error("error inserting new media", "err", err)
					return res.SetText(textErr), err
				}

				return res.SetText(textRemember), nil

			} else {
				if inc.RememberKind == KindCalloutCmd {
					c := &db.Callout{
						GID:     gid,
						Callout: inc.RememberAbraxas,
						Text:    inc.RememberRest,
					}

					err := db.InsertCallout(ctx, c)
					if err != nil {
						slog.Error("error inserting callout.",
							"callout", inc.RememberAbraxas,
							"rest", inc.RememberRest,
						)
						return res.SetText(textErr), nil
					}
					return res.SetText(textRemember), nil
				}

				if inc.RememberKind == KindTriggerCmd {
					a := &db.Abraxas{
						GID:     gid,
						Abraxas: inc.RememberAbraxas,
					}

					if rePhoto.MatchString(inc.RememberRest) {
						a.Kind = "photo"
					} else if reVideo.MatchString(inc.RememberRest) {
						a.Kind = "video"
					} else {
						a.Kind = "photo"
					}

					if err := db.InsertAbraxas(ctx, a); err != nil {
						return res, err
					}

					return res.SetText(textRemember), nil
				}
			}

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
					slog.Error("error deleting media", "err", err)

					res.SetText(textErr)
				} else {
					res.SetText(textForget)
				}
			}

			res.SendMessage().
				SetText(textForget)

		case CmdLink:
			switch inc.Operation {
			case OpAdd:
				slog.Debug("CmdLink#OpAdd")

				url, err := getURL(p)
				if err != nil {
					return res.SendMessage().
						SetText(textErr), err
				}

				link := &db.Link{
					GID:  gid,
					URL:  url,
					Text: text,
				}

				if err := db.InsertLink(ctx, link); err != nil {
					slog.Error("error inserting link", "err", err)

					return res.SendMessage().
						SetText(textErr), err
				}

				return res.SendMessage().
					SetText(textRemember), nil

			case OpRem:
				slog.Debug("CmdLink#OpRem")

				url, err := getURL(p)
				if err != nil {
					return res.SendMessage().
						SetText(textErr), err
				}

				link := &db.Link{
					GID: gid,
					URL: url,
				}

				if err := db.DeleteLink(ctx, link); err != nil {
					slog.Error("error deleting link", "err", err)

					return res.SendMessage().
						SetText(textErr), err
				}

				return res.SendMessage().
					SetText(textForget), nil

			default:
				slog.Debug("CmdLink#default")

				links, err := db.SearchLinks(ctx, gid, text)
				if err != nil {
					slog.Error("error searching links", "err", err)

					return res.SetText(textErr), err
				} else if len(links) == 0 {
					return res.SetText(text404), nil
				}
				var buttons []telegram.URLButton
				for _, v := range links {
					buttons = append(buttons, telegram.URLButton{
						URL:  v.URL,
						Text: v.Text,
					})
				}
				return res.SetText("links:").
					SetLinks(buttons), nil
			}

		case CmdDice:
			dies := dicer.New(inc.Rest)
			t := ""
			for _, die := range dies {
				t += die.Markdown()
			}

			return res.SendDice().
				SetText(t), nil
		}
	}

	return res.Empty(), nil
}
