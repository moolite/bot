package statistics

import (
	"context"

	"github.com/moolite/bot/pkg/tg"
)

func BotMiddleware(ctx context.Context, update *tg.Update) *tg.Update {
	return update
}
