// Package tg is a lightweight custom framework for building Telegram bots via webhooks.
//
// # Bot Lifecycle
//
// Create a bot with a token and webhook URL:
//
//	b, err := tg.New("BOT_TOKEN", "https://yourdomain.com/t/WEBHOOK_SECRET")
//
// Register handlers for different update types:
//
//	b.RegisterHandlers(
//	    &tg.UpdateHandler{Type: tg.UPD_STARTSWITH, Param: "/start", Fn: HandleStart},
//	    &tg.UpdateHandler{Type: tg.UPD_CALLBACK, Fn: HandleCallback},
//	)
//
// Use the HTTP handler in your chi router:
//
//	r.Post("/t/{apikey}", b.HttpHandler(logger))
//
// # Handler Types
//
// UpdateHandlerType defines how handlers match updates:
//
//	UPD_STARTSWITH - matches messages starting with a prefix
//	UPD_CONTAINS   - matches messages containing a substring
//	UPD_REGEXP     - matches messages matching a regex pattern
//	UPD_CALLBACK   - matches callback queries from inline keyboards
//	UPD_MEDIAPHOTO - matches messages containing photos
//	UPD_MEDIAVIDEO - matches messages containing videos
//	UPD_MEDIADOCUMENT - matches messages containing documents
//	UPD_WILDCARD   - matches all messages (fallback)
//	UPD_MENTION    - matches messages with @mention entities
//
// Handlers are executed in registration order; the first match wins.
// Register a default message handler with RegisterMessageHandler.
//
// # Sending Messages
//
// Build a Sendable with the desired method and parameters:
//
//	s := &tg.Sendable{
//	    Method: tg.MethodSendMessage,
//	    ChatID: 123456,
//	    Text:   "Hello!",
//	}
//	b.Send(ctx, s)
//
// Or use helper functions like SendableSetMessageReaction for common operations.
//
// # Types
//
// Core data types mirror Telegram's API schema with memory-optimized field ordering
// (fixed-size types first, strings/pointers last):
//
//	User, Chat, Message     - identity and conversation primitives
//	Audio, PhotoSize, Video - media descriptors
//	Document, Animation, Voice, VideoNote - file types
//	Location, Venue, Contact - shared content types
//	MessageEntity           - text formatting (bold, italic, URLs, etc.)
//	InlineKeyboardMarkup    - inline keyboard buttons
//	Update                  - incoming webhook update
//	Sendable                - outbound API request
//
// # Constants
//
// Entity type constants (ENTITY_BOLD, ENTITY_URL, etc.) for parsing message entities.
// Emoji constants (EMOJI_FIRE, EMOJI_HEART, etc.) for reactions and responses.
// Method constants (MethodSendMessage, MethodSendPhoto, etc.) for Sendable.Method.
package tg
