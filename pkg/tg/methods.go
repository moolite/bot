package tg

import "context"

const (
	AnswerCallbackQueryMethod = `answerCallbackQuery`
	SendMessageMethod         = `sendMessage`
	SetMessageReactionMethod  = `setMessageReaction`
	SetMyCommandsMethod       = `setMyCommands`
)

type Answer interface{}

func (b *Bot) SendSendable(ctx context.Context, s *Sendable) (*Answer, error) {
	if s.Method == "" {
		return nil, &ErrNoMethod{fn: "SendSendable"}
	}

	var res *Answer
	err := b.SendRaw(ctx, s.Method, s, res)
	return res, err
}

func (b *Bot) SetMyCommands(ctx context.Context, cmds *SetMyCommandsParams) (any, error) {
	var res *Answer
	err := b.SendRaw(ctx, SetMyCommandsMethod, cmds, res)
	return res, err
}

func SetMessageReaction(chatId, messageId int64, emoji ...string) *Sendable {
	reactions := make([]ReactionType, len(emoji))
	for idx, e := range emoji {
		reactions[idx] = ReactionType{Type: "emoji", Emoji: e}
	}

	return &Sendable{
		Method:    SetMessageReactionMethod,
		MessageID: messageId,
		ChatID:    chatId,
		Reaction:  reactions,
	}
}
