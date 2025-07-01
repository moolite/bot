package core

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"
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
			// Search
			{Command: "search", Description: "Search media files"},
			// Top
			{Command: "top", Description: "Top 10 media files"},
			// grumpyness
			{Command: "grumpyness", Description: "Grumpyness measuring instrument"},
		},
	})
	return err
}

func registerBotHandlers(_ context.Context, b *tg.Bot) {
	b.RegisterHandlers(
		&tg.UpdateHandler{
			Type:  tg.UPD_CALLBACK,
			Param: "",
			Fn:    OnCallback,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   `/search`,
			Aliases: []string{"/pupy", "/s"},
			Fn:      MediaSearchCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/top",
			Aliases: []string{"/t"},
			Fn:      MediaToptenCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/pot",
			Aliases: []string{"/p"},
			Fn:      MediaBottomtenCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/remember",
			Aliases: []string{"/ricorda", "/add", "/touch"},
			Fn:      MediaRememberCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/forget",
			Aliases: []string{"/dimentica", "/del", "/rm"},
			Fn:      MediaForgetCommand,
		},
		&tg.UpdateHandler{
			Type:    tg.UPD_STARTSWITH,
			Param:   "/abraxas",
			Aliases: []string{"/trigger", "/abx"},
			Fn:      AbraxasCommand,
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
			Type:    tg.UPD_STARTSWITH,
			Param:   "/grumpyness",
			Aliases: []string{"/grumpy", "/g"},
			Fn:      GrumpyCommand,
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

const (
	CB_DATANULL int64 = iota - 1
	CB_ERROR
	CB_MEDIA_UP
	CB_MEDIA_DOWN
	CB_MEDIA_SHOW
	CB_MEDIA_SEARCH_MORE
	CB_MEDIA_SEARCH_LESS
)

func formatCallbackData(act, data int64) string {
	return fmt.Sprintf("%x:%x", act, data)
}

func parseCallbackData(raw string) (int64, int64) {
	if raw == "" {
		return CB_ERROR, CB_DATANULL
	}

	data := strings.Split(raw, `:`)
	if len(data) == 1 {
		data = append(data, `-1`)
	}

	slog.Debug("parseCallbackData", "raw", raw, "parts", data)

	acti64, err := strconv.ParseInt(data[0], 16, 64)
	if err != nil {
		return CB_ERROR, CB_DATANULL
	}

	datai64, err := strconv.ParseInt(data[1], 16, 64)
	if err != nil {
		slog.Error("error in strconv", "err", err, "datai64", datai64, "data", data[1])
		return acti64, CB_DATANULL
	}

	return acti64, datai64
}

func OnCallback(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	action, data := parseCallbackData(update.CallbackQuery.Data)
	var err error
	switch action {
	case CB_MEDIA_UP:
		err = voteMedia(ctx, b, update, action, data)
	case CB_MEDIA_DOWN:
		err = voteMedia(ctx, b, update, action, data)
	case CB_MEDIA_SHOW:
		err = showMedia(ctx, b, update, action, data)
	case CB_MEDIA_SEARCH_LESS:
		fallthrough
	case CB_MEDIA_SEARCH_MORE:
		err = MediaSearchOffsetCallback(ctx, b, update, action, data)
	}

	return &tg.Sendable{
		Method:          tg.MethodAnswerCallbackQuery,
		CallbackQueryID: update.CallbackQuery.ID,
	}, err
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
	head, rest := utils.SplitMessageWords(cmp.Or(update.Message.Text, update.Message.Caption))
	head = strings.ToLower(head)
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

	if len(rest) > 0 {
		if err := db.SearchRandomMedia(ctx, media, rest); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				slog.Error("media search, nothing found, continuing")
				// continue...
			} else {
				slog.Error("error in db.SelectRandomMedia()", "err", err)
				return nil, err
			}
		} else {
			return mediaSendable(update.Message.Chat.ID, media), nil
		}
	}

	if err := db.SelectRandomMedia(ctx, media); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("random media not found")
			return nil, err
		}

		slog.Error("error in db.SelectRandomMedia()", "err", err)
		return nil, err
	}

	return mediaSendable(update.Message.Chat.ID, media), nil
}

//
// Dice
//

// DiceCommand handles the `/dice <dice>[ <dice>]` command
func DiceCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	head, rest := utils.SplitMessageWords(update.Message.Text)
	slog.Debug("DiceCommand", "head", head, "rest", rest)

	if rest == "" {
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

	if rest == "" {
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

	if callout.Text == "" {
		return nil, nil
	}

	text := strings.ReplaceAll(callout.Text, "%", fmt.Sprintf("<b>%s</b>", rest))

	return &tg.Sendable{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: "html",
		Method:    tg.MethodSendMessage,
	}, nil
}

// Media
func MediaSearchCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	_, rest := utils.SplitMessageWords(update.Message.Text)

	items, err := db.SearchMedia(ctx, update.Message.Chat.ID, rest, 0)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return tg.SendableSetMessageReaction(update, tg.EMOJI_KO), nil
	}

	keyboard := mediaSearchKeyboard(items, 0)
	snd := &tg.Sendable{
		ChatID:      update.Message.Chat.ID,
		Text:        rest,
		ParseMode:   "html",
		Method:      tg.MethodSendMessage,
		ReplyMarkup: keyboard,
	}

	slog.Debug("update in OnMessage", "update", update)

	return snd, nil
}

func MediaSearchOffsetCallback(ctx context.Context, b *tg.Bot, update *tg.Update, act, offset int64) error {
	if update.CallbackQuery.Message == nil {
		slog.Error("MediaSearchOffsetCallback(): CallbackQuery doesn't carry Message", "update", slog.AnyValue(update))
		return nil
	}

	term := update.CallbackQuery.Message.Text
	items, err := db.SearchMedia(ctx, update.CallbackQuery.Message.Chat.ID, term, int(offset))
	if err != nil {
		slog.Error("error in SearchMedia", "err", slog.AnyValue(err))
		return err
	}

	if len(items) == 0 {
		return nil
	}

	keyboard := mediaSearchKeyboard(items, offset)
	snd := &tg.Sendable{
		ChatID:      update.CallbackQuery.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.MessageID,
		Method:      tg.MethodEditMessageReplyMarkup,
		ReplyMarkup: keyboard,
	}

	slog.Debug("sending MediaSearchOffsetCallback", "snd", slog.AnyValue(snd))
	if res, err := b.Send(ctx, snd); err != nil {
		slog.Error("MediaSearchOffsetCallback() error", "act", act, "offset", offset, "res", slog.AnyValue(res))
		return err
	}

	return nil
}

func mediaSearchKeyboard(items []db.Media, offset int64) *tg.InlineKeyboardMarkup {
	k := &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{},
	}

	for row := range slices.Chunk(items, 3) {
		kbRow := make([]tg.InlineKeyboardButton, len(row))
		for idx, item := range row {
			kbRow[idx] = tg.InlineKeyboardButton{
				Text:         item.Description,
				CallbackData: formatCallbackData(CB_MEDIA_SHOW, item.RowID),
			}
		}

		k.InlineKeyboard = append(k.InlineKeyboard, kbRow)
	}

	backOffset := offset - 6
	if backOffset < 0 {
		backOffset = 0
	}
	k.InlineKeyboard = append(k.InlineKeyboard, []tg.InlineKeyboardButton{
		{Text: "<<", CallbackData: formatCallbackData(CB_MEDIA_SEARCH_LESS, backOffset)},
		{Text: ">>", CallbackData: formatCallbackData(CB_MEDIA_SEARCH_LESS, offset+6)},
	})

	return k
}

func MediaToptenCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	media, err := db.SelectMediaTop(ctx, update.Message.Chat.ID, 10)
	if err != nil {
		return nil, err
	}

	snd := mediaCollection(update.Message.Chat.ID, media[0].Description, media)
	return snd, nil
}

func MediaBottomtenCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	media, err := db.SelectMediaBottom(ctx, update.Message.Chat.ID, 10)
	if err != nil {
		return nil, err
	}

	snd := mediaCollection(update.Message.Chat.ID, media[0].Description, media)
	return snd, nil
}

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

	_, rest := utils.SplitMessageWords(cmp.Or(update.Message.Text, update.Message.Caption))
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
	ok := tg.SendableSetMessageReaction(update, tg.EMOJI_OK)
	ko := tg.SendableSetMessageReaction(update, tg.EMOJI_KO)

	data := ""
	if isPhoto(update) {
		data = getPhotoFileId(update)
	} else if isVideo(update) {
		data = getVideoFileID(update)
	} else {
		return ko, nil
	}

	media := &db.Media{
		GID:  update.Message.Chat.ID,
		Data: data,
	}

	if err := db.DeleteMedia(ctx, media); err != nil {
		if errors.Is(err, db.ErrDelete) {
			return ko, nil
		}

		slog.Error("error in db.DeleteMedia()", "media", media, "err", err)
		return ko, err
	} else {
		return ok, nil
	}
}

func voteMedia(ctx context.Context, b *tg.Bot, update *tg.Update, action, data int64) error {
	// Avoid checking inaccessible messages
	if update.CallbackQuery.Message == nil {
		return nil
	}

	msg := update.CallbackQuery.Message

	media := &db.Media{RowID: data}
	if err := db.SelectOneMediaByRowID(ctx, media); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("media not found from query", "callback_id", update.CallbackQuery.ID, "media.RowID", data)
			return nil
		}

		slog.Error("error in db.SelectOneMediaByData", "GID", media.GID, "Data", media.Data, "RowID", data, "err", err)
		return err
	}

	slog.Debug("updating media", "score", media.Score, "action", action)

	if action == CB_MEDIA_UP {
		media.Score += 1
	} else {
		media.Score -= 1
	}

	if err := db.UpdateMediaScoreByRowID(ctx, media); err != nil {
		slog.Error("error in db.InsertMedia", "GID", media.GID, "Data", media.Data, "RowID", data, "err", err)
		return nil
	}

	// Update the Previous Message
	keyboard := mediaKeyboard(media.RowID, media.Score)
	snd := &tg.Sendable{
		Method:      tg.MethodEditMessageReplyMarkup,
		ChatID:      media.GID,
		MessageID:   msg.MessageID,
		ReplyMarkup: keyboard,
	}

	slog.Debug("editing message", "msg", snd)
	if res, err := b.Send(ctx, snd); err != nil {
		slog.Error("VoteMedia error in b.Send", "err", err, "res", res)
		return err
	}
	return nil
}

func showMedia(ctx context.Context, b *tg.Bot, update *tg.Update, _, data int64) error {
	m := &db.Media{
		RowID: data,
	}
	if err := db.SelectOneMediaByRowID(ctx, m); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Debug("ShowMedia error", "err", err, "rowid", data)
			return nil
		}
		slog.Error("ShowMedia error", "err", err, "rowid", data)
		return err
	}

	slog.Debug("ShowMedia", "m", m)

	var chatId int64
	if update.Message != nil {
		chatId = update.Message.Chat.ID
	} else if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		chatId = update.CallbackQuery.Message.Chat.ID
	}

	snd := mediaSendable(chatId, m)

	if res, err := b.Send(ctx, snd); err != nil {
		slog.Error("ShowMedia error", "err", err, "res", res)
		return err
	}

	return nil
}

func mediaSendable(gid int64, m *db.Media) *tg.Sendable {
	keyboard := mediaKeyboard(m.RowID, m.Score)

	snd := &tg.Sendable{
		ChatID:      gid,
		Caption:     m.Description,
		ReplyMarkup: keyboard,
	}

	if m.Kind == "photo" {
		snd.Method = tg.MethodSendPhoto
		snd.Photo = m.Data
		return snd
	}

	if m.Kind == "video" {
		snd.Method = tg.MethodSendVideo
		snd.Video = m.Data
		return snd
	}

	return snd
}

func mediaKeyboard(uid int64, score int) *tg.InlineKeyboardMarkup {
	heart := tg.EMOJI_HEART
	if score >= 5 && score < 10 {
		heart = tg.EMOJI_KISS
	} else if score >= 10 && score < 15 {
		heart = tg.EMOJI_FIRE
	} else if score >= 15 {
		heart = tg.EMOJI_HEARTSTRUCK
	}

	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{
				{Text: heart, CallbackData: formatCallbackData(CB_MEDIA_UP, uid)},
				{Text: fmt.Sprintf("%d", score), CallbackData: " "},
				{Text: tg.EMOJI_HEARTBREAK, CallbackData: formatCallbackData(CB_MEDIA_DOWN, uid)},
			},
		},
	}
}

func mediaCollection(gid int64, caption string, items []db.Media) *tg.Sendable {
	snd := &tg.Sendable{
		ChatID:  gid,
		Method:  tg.MethodSendMediaGroup,
		Media:   make([]tg.InputMedia, len(items)),
		Caption: caption,
	}

	for i, item := range items {
		snd.Media[i] = tg.InputMedia{Type: item.Kind, Media: item.Data, Caption: item.Description}
	}

	return snd
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

// Grumpy
func GrumpyCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
	_, rest := utils.SplitMessageWords(update.Message.Text)
	points, user := utils.SplitMessageWords(rest)
	slog.Debug("GrumpyCommand", "parts", rest, "points", points, "user", user)

	if points != "" && user != "" {
		pt, err := strconv.ParseInt(points, 10, 64)
		if err != nil {
			slog.Error("error parsing points", slog.Any("points", points), slog.Any("pt", pt), slog.Any("err", err))
			return tg.SendableSetMessageReaction(update, tg.EMOJI_KO), nil
		}

		s := &db.ChannelStats{
			GID:    update.Message.Chat.ID,
			UID:    user,
			Points: pt,
		}
		if err = db.InsertChannelStats(ctx, s); err != nil {
			slog.Error("GrumpyCommand() InsertChannelStats error", slog.Any("err", err))
			return nil, err
		}
	}

	// finally show grumpiness index
	stats, err := db.SelectChannelStats(ctx, update.Message.Chat.ID)
	if err != nil {
		slog.Error("GrumpyCommand() SelectChannelStats error", slog.Any("err", err))
		return nil, err
	}

	text := "<b>Grumpiness</b>\n"
	if len(stats) == 0 {
		text = `<i>grumpiness is empty</i>`
	} else {
		sstats, err := db.SelectChannelStatsStats(ctx, update.Message.Chat.ID)
		if err != nil {
			return nil, err
		}

		for _, stat := range stats {
			text += fmt.Sprintf("<i>%s</i>: %3d\n", stat.UID, stat.Points)
		}

		text += fmt.Sprintf("\n<b>%d</b>min, <b>%d</b>max, <b>%d</b>sum, <b>%d</b>avg",
			sstats.Min,
			sstats.Max,
			sstats.Sum,
			sstats.Avg,
		)
	}

	snd := &tg.Sendable{
		Method:    tg.MethodSendMessage,
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: "html",
	}
	return snd, nil
}
