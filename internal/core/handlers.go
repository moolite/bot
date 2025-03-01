package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/dicer"
	"github.com/moolite/bot/internal/utils"
	"github.com/moolite/bot/pkg/tg"
)

func setReaction(chatId, messageId int64, emoji string) *tg.Sendable {
	return &tg.Sendable{
		Method:    "setMessageReaction",
		ChatID:    chatId,
		MessageID: messageId,
		Reaction: []tg.ReactionType{{
			Type:  "emoji",
			Emoji: emoji,
		}},
	}
}

func registerCommands(ctx context.Context, b *tg.Bot) error {
	_, err := b.SetMyCommands(ctx, &tg.SetMyCommandsParams{
		Commands: []tg.BotCommand{
			// Remember
			{Command: "remember", Description: "Remember a photo or video with a description"},
			// Forget
			{Command: "forget", Description: "Forget a photo or video"},
			// Callout
			{Command: "callout", Description: "Create or delete a !callout"},
			// Dice
			{Command: "dice", Description: "Roll a dice"},
			// Backup
			{Command: "backup", Description: "Create an upload a new database backup for the current chat"},
		},
	})
	return err
}

func registerBotHandlers(_ context.Context, b *tg.Bot) {
	b.RegisterHandlers(
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/remember",
			Aliases: []string{"/ricorda", "/add", "/touch"},
			Fn:      MediaRememberCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/forget",
			Aliases: []string{"/dimentica", "/rm"},
			Fn:      MediaForgetCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/dice",
			Aliases: []string{"/d"},
			Fn:      DiceCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/callout",
			Aliases: []string{"/c", "/oh"},
			Fn:      CalloutCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/alias",
			Aliases: []string{"/a"},
			Fn:      AliasCommand,
		},
		&tg.UpdateHandler{
			Type:  tg.UPD_STARTSWITH,
			Param: "/",
			Fn:    AnyCommand,
		},
		&tg.UpdateHandler{
			Type:  tg.UPD_STARTSWITH,
			Param: "!",
			Fn:    CalloutMessage,
		},
		&tg.UpdateHandler{
			Type:  tg.UPD_WILDCARD,
			Param: "",
			Fn:    OnMessage,
		},
	).
		RegisterMessageHandler(OnMessage)
}

// Any Message
func OnMessage(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	return nil, nil
}

//
// Command Aliases
//

// AliasCommand handles the `/alias <from> <to>` command
func AliasCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	var alias string
	var target string
	if _, rest := utils.SplitMessageWords(update.Message.Text); rest != "" {
		alias, rest = utils.SplitMessageWords(rest)
		target = utils.FirstWord(rest)
	}

	if alias != "" && target != "" {
		return nil, nil
	}

	switch alias {
	case "remember":
		target = "remember"
	case "forget":
		target = "forget"
	case "dice":
		target = "dice"
	case "callout":
		target = "callout"
	default:
		return nil, nil
	}

	if err := db.InsertAlias(ctx, &db.Alias{Name: alias, Target: target}); err != nil {
		return nil, err
	}

	return tg.SetMessageReaction(update.Message.Chat.ID, update.Message.ID, tg.EMOJI_OK), nil
}

//
// Any Command
//

// AnyCommand handles any command sent to the chat, searches for aliases and dispatch the update accordingly
func AnyCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	head := utils.FirstWord(update.Message.Text)
	head = head[1:]

	alias := &db.Alias{
		Name: head,
	}
	if err := db.SelectAlias(ctx, alias); err != nil {
		slog.Error("AnyCommand error during SelectAlias", "err", err)
		return nil, err
	}

	switch alias.Target {
	case "remember":
		return MediaRememberCommand(ctx, b, update)
	case "forget":
		return MediaForgetCommand(ctx, b, update)
	case "dice":
		return DiceCommand(ctx, b, update)
	case "callout":
		return CalloutCommand(ctx, b, update)
	default:
		return nil, nil
	}
}

//
// Abraxas
//

func messageHasAbraxas(update *tg.Update) bool {
	if update.Message == nil || isMedia(update) {
		return false
	}

	head := utils.FirstWord(update.Message.Text)
	if len(head) < 3 {
		return false
	}

	return true
}

// AbraxasHandler message text like: `abraxas ...`
func AbraxasHandler(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	head := utils.FirstWord(update.Message.Text)
	slog.Debug("AbraxasHandler", "head", head)

	chatId := update.Message.Chat.ID
	abraxas := &db.Abraxas{
		GID:     chatId,
		Abraxas: head,
	}

	if err := db.SelectOneAbraxasByAbraxas(ctx, abraxas); err != nil || len(abraxas.Kind) == 0 {
		slog.Error("error in db.SelectOneAbraxasByAbraxas()")
		return nil, err
	}

	media := &db.Media{
		GID:  chatId,
		Kind: abraxas.Kind,
	}
	if err := db.SelectRandomMedia(ctx, media); err != nil {
		slog.Error("error in db.SelectRandomMedia()")
		return nil, err
	}

	keyboard := mediaKeyboard(media.Score)
	keyboardJson, err := json.Marshal(keyboard)
	if err != nil {
		return nil, err
	}

	if media.Kind == "photo" {
		return &tg.Sendable{
			ChatID:      update.Message.Chat.ID,
			Photo:       media.Data,
			Caption:     media.Description,
			ParseMode:   "html",
			ReplyMarkup: string(keyboardJson),
		}, nil
	}

	if media.Kind == "video" {
		return &tg.Sendable{
			ChatID:      update.Message.Chat.ID,
			Video:       media.Data,
			Caption:     media.Description,
			ParseMode:   "html",
			ReplyMarkup: string(keyboardJson),
		}, nil
	}

	return nil, nil
}

//
// Dice
//

// DiceCommand handles the `/dice <dice>[ <dice>]` command
func DiceCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	head, rest := utils.SplitMessageWords(update.Message.Text)
	slog.Debug("DiceCommand", "head", head, "rest", rest)

	if len(rest) == 0 {
		return nil, nil
	}

	dies := dicer.New(rest)
	if len(dies) == 0 {
		return nil, nil
	}

	message := strings.Builder{}
	for idx, d := range dies {
		if idx > 0 {
			message.WriteString("\n---\n")
		}
		message.WriteString(d.HTML())
	}

	return &tg.Sendable{
		ChatID:    update.Message.Chat.ID,
		Text:      message.String(),
		ParseMode: "html",
		Method:    "sendMessage",
	}, nil
}

//
// Callouts
//

// CalloutCommand handles `/callout <name> <rest>`
func CalloutCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	_, rest := utils.SplitMessageWords(update.Message.Text)
	callout, rest := utils.SplitMessageWords(rest)
	slog.Debug("CalloutCommand", "callout", callout, "rest", rest)

	if len(callout) <= 3 || strings.Contains(callout, "%") {
		return nil, nil
	}

	c := &db.Callout{
		Callout: callout,
		Text:    rest,
		GID:     update.Message.Chat.ID,
	}

	if err := db.InsertCallout(ctx, c); err != nil {
		if err.Error() == "sql: no rows in result set" { // FIXME this should be a wrapped Error
			return setReaction(update.Message.Chat.ID, update.Message.ID, tg.EMOJI_KO), nil
		}
		return nil, err
	} else {
		return setReaction(update.Message.Chat.ID, update.Message.ID, tg.EMOJI_OK), nil
	}
}

func CalloutMessage(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	head, rest := utils.SplitMessageWords(update.Message.Text)
	head = head[1:] // remove '!'
	slog.Debug("CalloutMessage", "head", head, "rest", rest)

	if len(rest) == 0 {
		return nil, nil
	}

	callout := &db.Callout{
		GID:     update.Message.Chat.ID,
		Callout: head,
	}
	if err := db.SelectOneCallout(ctx, callout); err != nil {
		slog.Debug("error", "err", err)
		if strings.HasPrefix(err.Error(), "asd") {
			return setReaction(update.Message.Chat.ID, update.Message.ID, tg.EMOJI_KO), err
		}
	}

	if len(callout.Text) == 0 {
		return nil, nil
	}

	text := strings.ReplaceAll(callout.Text, "%", fmt.Sprintf("<b>%s</b>", rest))

	return &tg.Sendable{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: "html",
		Method:    "sendMessage",
	}, nil
}

//
// Media
//

func MediaRememberCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	kind := ""
	data := ""
	if isPhoto(update) {
		kind = "photo"
		data = update.Message.Photo[0].FileID
	} else if isVideo(update) {
		kind = "video"
		data = update.Message.Video.FileID
	} else {
		return nil, nil
	}

	_, rest := utils.SplitMessageWords(update.Message.Text)
	m := &db.Media{
		GID:         update.Message.Chat.ID,
		Data:        data,
		Kind:        kind,
		Description: rest,
		Score:       0,
	}

	if err := db.InsertMedia(ctx, m); err != nil {
		slog.Error("error in db.InsertMedia", "err", err)
	}
	return setReaction(update.Message.Chat.ID, update.Message.ID, tg.EMOJI_OK), nil
}

func MediaForgetCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	data := ""
	if isPhoto(update) {
		data = update.Message.Photo[0].FileID
	} else if isVideo(update) {
		data = update.Message.Video.FileID
	} else {
		return setReaction(update.Message.Chat.ID, update.Message.ID, tg.EMOJI_KO), nil
	}

	media := &db.Media{
		GID:  update.Message.Chat.ID,
		Data: data,
	}

	if err := db.DeleteMedia(ctx, media); err != nil {
		slog.Error("error in db.DeleteMedia()", "media", media, "err", err)
		return setReaction(update.Message.Chat.ID, update.Message.ID, tg.EMOJI_KO), err
	} else {
		return setReaction(update.Message.Chat.ID, update.Message.ID, tg.EMOJI_OK), nil
	}
}

const (
	MEDIA_UP   string = "media:up"
	MEDIA_DOWN string = "media:down"
)

func VoteMedia(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	// Avoid checking inaccessible messages
	if update.CallbackQuery.Message == nil {
		return nil, nil
	}

	msg := update.CallbackQuery.Message
	query := update.CallbackQuery.Data
	id := update.CallbackQuery.ID

	data := ""
	if msg.Video != nil {
		data = msg.Video.FileID
	}

	if len(msg.Photo) > 0 {
		data = msg.Photo[0].FileID
	}

	media := &db.Media{
		GID:  msg.Chat.ID,
		Data: data,
	}
	if err := db.SelectOneMediaByData(ctx, media); err != nil {
		slog.Error("error in db.SelectOneMediaByData", "GID", media.GID, "Data", media.Data, "err", err)
		return nil, err
	}

	if query == MEDIA_UP {
		media.Score += 1
	} else {
		media.Score -= 1
	}

	if err := db.InsertMedia(ctx, media); err != nil {
		slog.Error("error in db.InsertMedia", "GID", media.GID, "Data", media.Data, "err", err)
	}

	kb := mediaKeyboard(media.Score)
	keyboard, err := json.Marshal(kb)
	return &tg.Sendable{
		Method:          tg.AnswerCallbackQueryMethod,
		ChatID:          msg.Chat.ID,
		MessageID:       msg.ID,
		CallbackQueryID: id,
		ReplyMarkup:     string(keyboard),
	}, err
}

func mediaKeyboard(score int) *tg.InlineKeyboardMarkup {
	heart := tg.EMOJI_HEART
	if score > 10 {
		heart = tg.EMOJI_HEARTSTRUCK
	}

	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{
				{Text: heart, CallbackData: MEDIA_UP},
				{Text: fmt.Sprintf("<b><i>%d</i></b>", score)},
				{Text: tg.EMOJI_HEARTBREAK, CallbackData: MEDIA_DOWN},
			},
		},
	}
}

func isPhoto(update *tg.Update) bool {
	return update.Message != nil && len(update.Message.Photo) > 0
}

func isVideo(update *tg.Update) bool {
	return update.Message != nil && update.Message.Video != nil
}

func isMedia(update *tg.Update) bool {
	return isPhoto(update) || isVideo(update)
}
