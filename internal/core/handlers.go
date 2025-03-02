package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/dicer"
	"github.com/moolite/bot/internal/utils"
	"github.com/moolite/bot/pkg/tg"
)

func registerCommands(ctx context.Context, b *tg.Bot) error {
	_, err := b.SetMyCommands(ctx, &tg.SetMyCommandsParams{
		Commands: []tg.BotCommand{
			// Remember
			{Command: "remember", Description: "Remember a photo or video with a description"},
			// abraxas
			{Command: "abraxas", Description: "Manage a trigger. /abraxas <add|rm> <word> [photo|video]"},
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
			Type:  tg.UPD_CALLBACK,
			Param: "",
			Fn:    VoteMedia,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/remember",
			Aliases: []string{"/ricorda", "/add", "/touch"},
			Fn:      MediaRememberCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/abraxas",
			Aliases: []string{"/trigger", "/abx"},
			Fn:      AbraxasCommand,
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
	)

	b.RegisterMessageHandler(OnMessage)

}

// Any Message
func OnMessage(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	if update.Message != nil && update.Message.Text != "" {
		if snd, err := AbraxasHandler(ctx, b, update); err != nil {
			return nil, err
		} else if snd != nil {
			return snd, nil
		}
	}

	slog.Debug("update in OnMessage", "update", update)

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

	return tg.SendableSetMessageReaction(update, tg.EMOJI_OK), nil
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

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

// AbraxasCommand handles commands for abraxas (add, remove)
func AbraxasCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	slog.Error("AbraxasCommand is not implemented")
	return nil, nil
}

// AbraxasHandler handles message text like: `abraxas ...`
func AbraxasHandler(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	head := strings.ToLower(utils.FirstWord(update.Message.Text))
	slog.Debug("AbraxasHandler", "head", head)

	chatId := update.Message.Chat.ID
	abraxas := &db.Abraxas{
		GID:     chatId,
		Abraxas: head,
	}

	if err := db.SelectOneAbraxasByAbraxas(ctx, abraxas); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		slog.Error("error in db.SelectOneAbraxasByAbraxas()")
		return nil, err
	}

	media := &db.Media{
		GID:  chatId,
		Kind: abraxas.Kind,
	}
	if err := db.SelectRandomMedia(ctx, media); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("random media not found")
			return nil, err
		}

		slog.Error("error in db.SelectRandomMedia()")
		return nil, err
	}

	keyboard := mediaKeyboard(media.RowID, media.Score)

	snd := &tg.Sendable{
		ChatID:      update.Message.Chat.ID,
		Caption:     media.Description,
		ReplyMarkup: keyboard,
	}

	if media.Kind == "photo" {
		snd.Method = tg.MethodSendPhoto
		snd.Photo = media.Data
		return snd, nil
	}

	if media.Kind == "video" {
		snd.Method = tg.MethodSendVideo
		snd.Video = media.Data
		return snd, nil
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
			return tg.SendableSetMessageReaction(update, tg.EMOJI_KO), nil
		}
		return nil, err
	} else {
		return tg.SendableSetMessageReaction(update, tg.EMOJI_OK), nil
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		slog.Debug("error", "err", err)
		if strings.HasPrefix(err.Error(), "asd") {
			return tg.SendableSetMessageReaction(update, tg.EMOJI_KO), err
		}
		return nil, err
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
		data = getPhotoFileId(update)
	} else if isVideo(update) {
		kind = "video"
		data = getVideoFileID(update)
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

	return tg.SendableSetMessageReaction(update, tg.EMOJI_OK), nil
}

func MediaForgetCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	data := ""
	if isPhoto(update) {
		data = update.Message.Photo[0].FileID
	} else if isVideo(update) {
		data = update.Message.Video.FileID
	} else {
		return tg.SendableSetMessageReaction(update, tg.EMOJI_KO), nil
	}

	media := &db.Media{
		GID:  update.Message.Chat.ID,
		Data: data,
	}

	if err := db.DeleteMedia(ctx, media); err != nil {
		slog.Error("error in db.DeleteMedia()", "media", media, "err", err)
		return tg.SendableSetMessageReaction(update, tg.EMOJI_KO), err
	} else {
		return tg.SendableSetMessageReaction(update, tg.EMOJI_OK), nil
	}
}

const (
	MEDIA_UP   = `media:up`
	MEDIA_DOWN = `media:pu`
)

func VoteMedia(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	// Avoid checking inaccessible messages
	if update.CallbackQuery.Message == nil {
		return nil, nil
	}

	msg := update.CallbackQuery.Message
	id := update.CallbackQuery.ID
	answer := &tg.Sendable{
		Method:          tg.MethodAnswerCallbackQuery,
		CallbackQueryID: id,
	}

	callbackData := strings.Split(update.CallbackQuery.Data, `|`)
	if len(callbackData) < 2 {
		return answer, nil
	}

	query := callbackData[0]
	rowid, err := strconv.ParseInt(callbackData[1], 10, 64)
	if err != nil {
		slog.Debug("wrong CallbackQuery.Data data", "err", err, "callbackData", callbackData)
		return nil, nil
	}

	media := &db.Media{RowID: rowid}
	if err := db.SelectOneMediaByRowID(ctx, media); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("media not found from query", "callback_id", update.CallbackQuery.ID, "media.RowID", rowid)
			return answer, nil
		}

		slog.Error("error in db.SelectOneMediaByData", "GID", media.GID, "Data", media.Data, "err", err)
		return answer, err
	}

	slog.Debug("updating media", "score", media.Score, "query", query)

	if query == MEDIA_UP {
		media.Score += 1
	} else {
		media.Score -= 1
	}

	if err := db.UpdateMediaScoreByRowID(ctx, media); err != nil {
		slog.Error("error in db.InsertMedia", "GID", media.GID, "Data", media.Data, "err", err)
		return answer, nil
	}

	// Update the Previous Message
	keyboard := mediaKeyboard(media.RowID, media.Score)
	snd := &tg.Sendable{
		Method:      tg.MethodEditMessageReplyMarkup,
		ChatID:      msg.Chat.ID,
		MessageID:   msg.MessageID,
		ReplyMarkup: keyboard,
	}
	slog.Debug("editing message", "msg", snd)
	res := &tg.RawResult{}
	if err := b.Send(ctx, snd, res); err != nil {
		slog.Error("error sending editMessageReplyMarkup", "err", err, "res", res)
	}

	// NOTE: always answer with a answerCallbackQuery https://core.telegram.org/bots/api#answercallbackquery
	return answer, nil
}

func mediaKeyboard(uid int64, score int) *tg.InlineKeyboardMarkup {
	heart := tg.EMOJI_HEART
	if score > 10 {
		heart = tg.EMOJI_HEARTSTRUCK
	}

	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{
				{Text: heart, CallbackData: fmt.Sprintf("%s|%d", MEDIA_UP, uid)},
				{Text: fmt.Sprintf("%d", score), CallbackData: " "},
				{Text: tg.EMOJI_HEARTBREAK, CallbackData: fmt.Sprintf("%s|%d", MEDIA_DOWN, uid)},
			},
		},
	}
}

func isPhoto(update *tg.Update) bool {
	if update.Message == nil {
		return false
	}

	return len(update.Message.Photo) > 0 || update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.Photo != nil
}

func getPhotoFileId(update *tg.Update) string {
	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.Photo != nil {
		return update.Message.ReplyToMessage.Photo[0].FileID
	}

	return update.Message.Photo[0].FileID
}

func isVideo(update *tg.Update) bool {
	if update.Message == nil {
		return false
	}

	return update.Message.Video != nil || update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.Video != nil
}

func getVideoFileID(update *tg.Update) string {
	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.Video != nil {
		return update.Message.ReplyToMessage.Video.FileID
	}

	return update.Message.Video.FileID
}

func isMedia(update *tg.Update) bool {
	return isPhoto(update) || isVideo(update)
}
