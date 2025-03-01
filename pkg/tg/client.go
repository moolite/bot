package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"time"
)

func (bot *Bot) SendRaw(ctx context.Context, method string, data, results any) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	uri := path.Join(bot.URL.String(), method)

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	res, err := client.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resBody, results); err != nil {
		return err
	}

	return nil
}

func (b *Bot) Send(ctx context.Context, s *Sendable, res any) error {
	return b.SendRaw(ctx, s.Method, s, res)
}
