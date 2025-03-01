package tg

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
)

type BotError struct {
	function string
	cause    string
}

func (e *BotError) Error() string {
	return fmt.Sprintf("error in %s: %s", e.function, e.cause)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MiddlewareFn func(ctx context.Context, update *Update) *Update

type UpdateHandlerType int

const (
	UPD_STARTSWITH UpdateHandlerType = iota
	UPD_CONTAINS
	UPD_REGEXP
	UPD_MEDIAPHOTO
	UPD_MEDIAVIDEO
	UPD_MEDIADOCUMENT
	UPD_CALLBACK
	UPD_WILDCARD
)

type UpdateHandlerFn func(ctx context.Context, bot *Bot, update *Update) (*Sendable, error)

type UpdateHandler struct {
	Type    UpdateHandlerType
	Param   string
	Aliases []string
	Fn      UpdateHandlerFn

	rx *regexp.Regexp
}

type Bot struct {
	Token      string
	URL        *url.URL
	WebhookURL *url.URL

	WebhookSecretToken string
	handlers           []*UpdateHandler
	onMessageHandler   UpdateHandlerFn
	middlewares        []MiddlewareFn
	client             HTTPClient
	timeout            time.Duration
}

// New creates a new Bot
func New(token, webhookUrl string) (*Bot, error) {
	u, err := url.Parse(webhookUrl)
	if err != nil {
		return nil, err
	}

	return &Bot{
		Token:            token,
		URL:              u,
		timeout:          60 * time.Second,
		onMessageHandler: noopHandler,
	}, nil
}

func MustNew(token, webhookUrl string) *Bot {
	u, err := New(token, webhookUrl)
	if err != nil {
		panic(err)
	}
	return u
}

// New creates a new Bot using the given http client, the interface requires a `Do` method
func NewWithClient(token, webhookUrl string, client HTTPClient) (*Bot, error) {
	b, err := New(token, webhookUrl)
	if err != nil {
		return b, err
	}
	b.client = client
	return b, nil
}

// RegisterMesageHandler registers a default handler that catches all messages
func (b *Bot) RegisterMessageHandler(fn UpdateHandlerFn) *Bot {
	b.onMessageHandler = fn
	return b
}

func noopHandler(ctx context.Context, bot *Bot, update *Update) (*Sendable, error) { return nil, nil }

// RunHandlers cycle registered handlers will execute the first matching occurrence
func (b *Bot) RunHandlers(ctx context.Context, upd *Update) (*Sendable, error) {
	defaultRes, err := b.onMessageHandler(ctx, b, upd)
	if err != nil {
		return nil, err
	}

	for _, middleware := range b.middlewares {
		upd = middleware(ctx, upd)
	}

	text := ""
	if upd.Message != nil {
		text = cmp.Or(upd.Message.Caption, upd.Message.Text)
	}

	for _, handler := range b.handlers {
		switch handler.Type {
		case UPD_STARTSWITH:
			if strings.HasPrefix(text, handler.Param) {
				return handler.Fn(ctx, b, upd)
			}
			for _, alias := range handler.Aliases {
				if strings.HasPrefix(text, alias) {
					return handler.Fn(ctx, b, upd)
				}
			}
		case UPD_CONTAINS:
			if strings.Contains(text, handler.Param) {
				return handler.Fn(ctx, b, upd)
			}
			for _, alias := range handler.Aliases {
				if strings.Contains(text, alias) {
					return handler.Fn(ctx, b, upd)
				}
			}
		case UPD_REGEXP:
			if handler.rx.MatchString(text) {
				return handler.Fn(ctx, b, upd)
			}
		case UPD_MEDIADOCUMENT:
			if upd.Message.Document != nil {
				return handler.Fn(ctx, b, upd)
			}
		case UPD_MEDIAPHOTO:
			if upd.Message.Photo != nil {
				return handler.Fn(ctx, b, upd)
			}
		case UPD_MEDIAVIDEO:
			if upd.Message.Video != nil {
				return handler.Fn(ctx, b, upd)
			}
		case UPD_CALLBACK:
			if upd.CallbackQuery != nil {
				return handler.Fn(ctx, b, upd)
			}
		case UPD_WILDCARD:
			return handler.Fn(ctx, b, upd)
		}
	}

	return defaultRes, nil
}

func (b *Bot) RegisterMiddlewares(middlewares ...MiddlewareFn) *Bot {
	b.middlewares = append(b.middlewares, middlewares...)

	return b
}

func (b *Bot) RegisterHandlers(handlers ...*UpdateHandler) *Bot {
	for _, handler := range handlers {
		if handler.Type == UPD_REGEXP {
			handler.rx = regexp.MustCompile(handler.Param)
		}

		if !slices.Contains(b.handlers, handler) {
			b.handlers = append(b.handlers, handler)
		}
	}

	return b
}

func (b *Bot) HttpHandler(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeoutCause(
			r.Context(),
			b.timeout,
			&BotError{function: "HttpHandler", cause: "timeout"},
		)
		defer cancel()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			l.Error("error reading request", "err", err)
			return
		}

		upd := &Update{}
		if err := json.Unmarshal(data, upd); err != nil {
			l.Error("error unmarshaling request body", "err", err)
			return
		}

		snd, err := b.RunHandlers(ctx, upd)
		if err != nil {
			l.Error("error running handlers", "err", err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := json.Marshal(snd)
		if err != nil {
			l.Error("error marshaling sendable", "err", err)
			return
		}

		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(body); err != nil {
			l.Error("error writing body", "err", err)
		}
	}
}

func (b *Bot) SetWebhook(ctx context.Context) error {
	var res interface{}
	a := &ParamSetWebhook{
		URL: b.WebhookURL,
	}
	return b.SendRaw(ctx, "setWebhook", a, res)
}

func (b *Bot) DeleteWebhook(ctx context.Context, dropUpdates bool) error {
	var res interface{}
	a := &ParamDeleteWebhook{
		DropPendingUpdates: dropUpdates,
	}
	return b.SendRaw(ctx, "deleteWebhook", a, res)
}
