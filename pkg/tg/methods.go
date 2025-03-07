package tg

import (
	"cmp"
	"context"
	"fmt"
)

const (
	MethodAnswerCallbackQuery    = `answerCallbackQuery`
	MethodEditMessageReplyMarkup = `editMessageReplyMarkup`
	MethodEditMessageText        = `editMessageText`
	MethodSendMediaGroup         = `sendMediaGroup`
	MethodSendMessage            = `sendMessage`
	MethodSendPhoto              = `sendPhoto`
	MethodSendVideo              = `sendVideo`
	MethodSetMessageReaction     = `setMessageReaction`
	MethodSetMyCommands          = `setMyCommands`
)

func (b *Bot) SendSendable(ctx context.Context, s *Sendable) (*RawResult, error) {
	if s.Method == "" {
		return nil, &ErrNoMethod{fn: "SendSendable"}
	}

	res := &RawResult{}
	err := b.SendRaw(ctx, s.Method, s, res)
	return res, err
}

func (b *Bot) SetMyCommands(ctx context.Context, cmds *SetMyCommandsParams) (any, error) {
	res := &RawResult{}
	err := b.SendRaw(ctx, MethodSetMyCommands, cmds, res)
	return res, err
}

func SendableSetMessageReaction(upd *Update, emoji ...string) *Sendable {
	chatId := upd.Message.Chat.ID
	messageId := cmp.Or(upd.Message.ID, upd.Message.MessageID)
	reactions := make([]ReactionType, len(emoji))
	for idx, e := range emoji {
		reactions[idx] = ReactionType{Type: "emoji", Emoji: e}
	}

	return &Sendable{
		Method:    MethodSetMessageReaction,
		MessageID: messageId,
		ChatID:    chatId,
		Reaction:  reactions,
	}
}

func (b *Bot) SetMessageReaction(ctx context.Context, upd *Update, emoji ...string) error {
	body := SendableSetMessageReaction(upd, emoji...)
	res := &RawResult{}

	if err := b.SendRaw(ctx, MethodSetMessageReaction, body, res); err != nil {
		return fmt.Errorf("error when setting message reaction %v: %v", res, err)
	}

	return nil
}
