package tg

import "net/url"

// User is telegram user
type User struct {
	ID                      int64  `json:"id"`
	IsBot                   bool   `json:"is_bot"`
	FirstName               string `json:"first_name"`
	LastName                string `json:"last_name"`
	Username                string `json:"username"`
	LanguageCode            string `json:"language_code"`
	CanJoinGroups           bool   `json:"can_join_groups"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries"`
}

// ChatPhoto represents a chat photo
type ChatPhoto struct {
	SmallFileID       string `json:"small_file_id"`
	SmallFileUniqueID string `json:"small_file_unique_id"`
	BigFileID         string `json:"big_file_id"`
	BigFileUniqueID   string `json:"big_file_unique_id"`
}

// Chat represents a chat
type Chat struct {
	ID                          int64            `json:"id"`
	Type                        string           `json:"type,omitempty"`
	Title                       string           `json:"title,omitempty"`
	Username                    string           `json:"username,omitempty"`
	FirstName                   string           `json:"first_name,omitempty"`
	LastName                    string           `json:"last_name,omitempty"`
	Photo                       *ChatPhoto       `json:"photo,omitempty"`
	Description                 string           `json:"descritpion,omitempty"`
	InviteLink                  string           `json:"invite_link,omitempty"`
	PinnedMessage               *Message         `json:"pinned_message,omitempty"`
	Permissions                 *ChatPermissions `json:"permissions,omitempty"`
	StickerSetName              string           `json:"sticker_set_name,omitempty"`
	AllMembersAreAdministrators bool             `json:"all_members_are_administrators,omitempty"` // deprecated
	SlowModeDelay               int              `json:"slow_mode_delay,omitempty"`
	CanSetStickerSet            bool             `json:"can_set_sticker_set,omitempty"`
}

// ChatPermissions describes actions that a non-administrator user is allowed to take in a chat.
type ChatPermissions struct {
	CanSendMessages       bool `json:"can_send_messages,omitempty"`         // True, if the user is allowed to send text messages, contacts, locations and venues
	CanSendMediaMessages  bool `json:"can_send_media_messages,omitempty"`   // True, if the user is allowed to send audios, documents, photos, videos, video notes and voice notes, implies can_send_messages
	CanSendPolls          bool `json:"can_send_polls,omitempty"`            // True, if the user is allowed to send polls, implies can_send_messages
	CanSendOtherMessages  bool `json:"can_send_other_messages,omitempty"`   // True, if the user is allowed to send animations, games, stickers and use inline bots, implies can_send_media_messages
	CanAddWebPagePreviews bool `json:"can_add_web_page_previews,omitempty"` // True, if the user is allowed to add web page previews to their messages, implies can_send_media_messages
	CanChangeInfo         bool `json:"can_change_info,omitempty"`           // True, if the user is allowed to change the chat title, photo and other settings. Ignored in public supergroups
	CanInviteUsers        bool `json:"can_invite_users,omitempty"`          // True, if the user is allowed to invite new users to the chat
	CanPinMessages        bool `json:"can_pin_messages,omitempty"`          // True, if the user is allowed to pin messages. Ignored in public supergroups
}

// MessageEntity represents one special entity in a text message.
// For example, hashtags, usernames, URLs, etc.
type MessageEntity struct {
	Type     string `json:"type,omitempty"`
	Offset   int    `json:"offse,omitemptyt"`
	Length   int    `json:"lengt,omitemptyh"`
	URL      string `json:"ur,omitemptyl"`
	User     *User  `json:"use,omitemptyr"`
	Language string `json:"languag,omitemptye"`
}

// Audio represents an audio file to be treated as music by the Telegram clients
type Audio struct {
	FileID       string `json:"file_id,omitempty"`
	FileUniqueID string `json:"file_unique_id,omitempty"`
	Duration     int    `json:"duration,omitempty"`
	Performer    string `json:"performer,omitempty"`
	Title        string `json:"title,omitempty"`
	MIMEType     string `json:"mime_type,omitempty"`
	FileSize     int    `json:"file_size,omitempty"`
}

// PhotoSize represents one size of a photo or a file/sticker thumbnail.
type PhotoSize struct {
	FileID       string `json:"file_id,omitempty"`
	FileUniqueID string `json:"file_unique_id,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	FileSize     int    `json:"file_size,omitempty"`
}

// Document represents a general file
// (as opposed to photos, voice messages and audio files)
type Document struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Thumb        *PhotoSize `json:"thumb"`
	FileName     string     `json:"file_name"`
	MIMEType     string     `json:"mime_type"`
	FileSize     int        `json:"file_size"`
}

// Game represents a game. Use BotFather to create and edit games,
// their short names will act as unique identifiers.
type Game struct {
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	Photo        []*PhotoSize     `json:"photo"`
	Text         string           `json:"text"`
	TextEntities []*MessageEntity `json:"text_entities"`
	Animation    *Animation       `json:"animation"`
}

// Animation represents an animation file
// to be displayed in the message containing a game
type Animation struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Thumb        *PhotoSize `json:"thumb"`
	FileName     string     `json:"file_name"`
	MimeType     string     `json:"mime_type"`
	FileSize     int        `json:"file_size"`
}

// Sticker represents a sticker
type Sticker struct {
	FileID       string        `json:"file_id"`
	FileUniqueID string        `json:"file_unique_id"`
	Width        int           `json:"width"`
	Height       int           `json:"height"`
	IsAnimated   bool          `json:"is_animated"`
	Thumb        *PhotoSize    `json:"thumb"`
	Emoji        string        `json:"emoji"`
	MaskPosition *MaskPosition `json:"mask_position"`
	SetName      string        `json:"set_name"`
	FileSize     int           `json:"file_size"`
}

// MaskPosition describes the position on faces
// where a mask should be placed by default
type MaskPosition struct {
	Point  string  `json:"point"`
	XShift float32 `json:"x_shift"`
	YShift float32 `json:"y_shift"`
	Scale  float32 `json:"scale"`
}

// Video represents a video file
type Video struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	Duration     int        `json:"duration"`
	Thumbnail    *PhotoSize `json:"thumb"`
	MimeType     string     `json:"mime_type"`
	FileSize     int        `json:"file_size"`
}

// Voice represents a voice note
type Voice struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Duration     int    `json:"duration"`
	MimeType     string `json:"mime_type"`
	FileSize     int    `json:"file_size"`
}

// VideoNote represents a video message
type VideoNote struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Length       int        `json:"length"`
	Duration     int        `json:"duration"`
	Thumb        *PhotoSize `json:"thumb"`
	FileSize     int        `json:"file_size"`
}

// Contact represents a phone contact
type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	UserID      int    `json:"user_id"`
}

// Location represents a point on the map
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Venue represents a venue
type Venue struct {
	Location     Location `json:"location"`
	Title        string   `json:"title"`
	Address      string   `json:"address"`
	FoursquareID string   `json:"foursquare_id"`
}

// Invoice contains basic information about an invoice
type Invoice struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	StartParameter string `json:"start_parameter"`
	Currency       string `json:"currency"`
	TotalAmount    int    `json:"total_amount"`
}

// SuccessfulPayment contains basic information about a successful payment
type SuccessfulPayment struct {
	Currency                string     `json:"currency"`
	TotalAmount             int        `json:"total_amount"`
	InvoicePayload          string     `json:"invoice_payload"`
	ShippingOptionID        string     `json:"shipping_option_id"`
	OrderInfo               *OrderInfo `json:"order_info"`
	TelegramPaymentChargeID string     `json:"telegram_payment_charge_id"`
	ProviderPaymentChargeID string     `json:"provider_payment_charge_id"`
}

// OrderInfo represents information about an order
type OrderInfo struct {
	Name            string           `json:"name"`
	PhoneNumber     string           `json:"phone_number"`
	Email           string           `json:"email"`
	ShippingAddress *ShippingAddress `json:"shipping_address"`
}

// ShippingAddress represents a shipping address
type ShippingAddress struct {
	CountryCode string `json:"country_code"`
	State       string `json:"state"`
	City        string `json:"city"`
	StreetLine1 string `json:"street_line1"`
	StreetLine2 string `json:"street_line2"`
	PostCode    string `json:"post_code"`
}

// Message represents a message
type Message struct {
	ID                    int64                 `json:"id,omitempty"`
	MessageID             int64                 `json:"message_id,omitempty"`
	From                  *User                 `json:"from,omitemtpy"`
	Date                  int64                 `json:"date,omitempty"`
	Chat                  Chat                  `json:"chat"`
	ForwardFrom           *User                 `json:"forward_from,omitempty"`
	ForwardFromChat       *Chat                 `json:"forward_from_chat,omitempty"`
	ForwardFromMessageID  int                   `json:"forward_from_message_id,omitempty"`
	ForwardSignature      string                `json:"forward_signature,omitempty"`
	ForwardSenderName     string                `json:"forward_sender_name,omitempty"`
	ForwardDate           int64                 `json:"forward_date,omitempty"`
	ReplyToMessage        *Message              `json:"reply_to_message,omitempty"`
	EditDate              int64                 `json:"edit_date,omitempty"`
	MediaGroupID          string                `json:"media_group_id,omitempty"`
	AuthorSignature       string                `json:"author_signature,omitempty"`
	Text                  string                `json:"text,omitempty"`
	Entities              []*MessageEntity      `json:"entities,omitempty"`
	CaptionEntities       []*MessageEntity      `json:"caption_entities,omitempty"`
	Audio                 *Audio                `json:"audio,omitempty"`
	Document              *Document             `json:"document,omitempty"`
	Game                  *Game                 `json:"game,omitempty"`
	Photo                 []*PhotoSize          `json:"photo,omitempty"`
	Sticker               *Sticker              `json:"sticker,omitempty"`
	Video                 *Video                `json:"video,omitempty"`
	Voice                 *Voice                `json:"voice,omitempty"`
	VideoNote             *VideoNote            `json:"video_note,omitempty"`
	Caption               string                `json:"caption,omitempty"`
	Contact               *Contact              `json:"contact,omitempty"`
	Location              *Location             `json:"location,omitempty"`
	Venue                 *Venue                `json:"venue,omitempty"`
	Poll                  *Poll                 `json:"poll,omitempty"`
	Dice                  *Dice                 `json:"dice,omitempty"`
	NewChatMembers        []*User               `json:"new_chat_members,omitempty"`
	LeftChatMember        *User                 `json:"left_chat_member,omitempty"`
	NewChatTitle          string                `json:"new_chat_title,omitempty"`
	NewChatPhoto          []*PhotoSize          `json:"new_chat_photo,omitempty"`
	DeleteChatPhoto       bool                  `json:"delete_chat_photo,omitempty"`
	GroupChatCreated      bool                  `json:"group_chat_created,omitempty"`
	SupergroupChatCreated bool                  `json:"supergroup_chat_created,omitempty"`
	ChannelChatCreated    bool                  `json:"channel_chat_created,omitempty"`
	MigrateToChatID       int                   `json:"migrate_to_chat_id,omitempty"`
	MigrateFromChatID     int                   `json:"migrate_from_chat_id,omitempty"`
	PinnedMessage         *Message              `json:"pinned_message,omitempty"`
	Invoice               *Invoice              `json:"invoice,omitempty"`
	SuccessfulPayment     *SuccessfulPayment    `json:"successful_payment,omitempty"`
	ConnectedWebsite      string                `json:"connected_website,omitempty"`
	PassportData          *PassportData         `json:"passport_data,omitempty"`
	ReplyMarkup           *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// InlineKeyboardMarkup represents an inline keyboard that appears right next to the message it belongs to
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton represents one button of an inline keyboard
type InlineKeyboardButton struct {
	Text                         string    `json:"text"`
	URL                          string    `json:"url,omitempty"`
	LoginURL                     *LoginURL `json:"login_url,omitempty"`
	CallbackData                 string    `json:"callback_data,omitempty"`
	SwitchInlineQuery            *string   `json:"switch_inline_query,omitempty"`
	SwitchInlineQueryCurrentChat *string   `json:"switch_inline_query_current_chat,omitempty"`
}

// LoginURL is a property of InlineKeyboardButton for Seamless Login feature
type LoginURL struct {
	URL                string  `json:"url"`
	ForwardText        *string `json:"forward_text,omitempty"`
	BotUsername        *string `json:"bot_username,omitempty"`
	RequestWriteAccess *string `json:"request_write_access,omitempty"`
}

// InlineQuery represents an incoming inline query
type InlineQuery struct {
	ID       string    `json:"id"`
	From     *User     `json:"from"`
	Location *Location `json:"location"`
	Query    string    `json:"query"`
	Offset   string    `json:"offset"`
}

// ChosenInlineResult represents a result of an inline query
// that was chosen by the user and sent to their chat partner
type ChosenInlineResult struct {
	ResultID        string    `json:"result_id"`
	From            *User     `json:"from"`
	Location        *Location `json:"location"`
	InlineMessageID string    `json:"inline_message_id"`
	Query           string    `json:"query"`
}

// CallbackQuery represents an incoming callback query
// from a callback button in an inline keyboard
type CallbackQuery struct {
	ID              string   `json:"id"`
	From            *User    `json:"from"`
	Message         *Message `json:"message"`
	InlineMessageID string   `json:"inline_message_id"`
	ChatInstance    string   `json:"chat_instance"`
	Data            string   `json:"data"`
	GameShortName   string   `json:"game_short_name"`
}

// ShippingQuery contains information about an incoming shipping query
type ShippingQuery struct {
	ID              string           `json:"id"`
	From            *User            `json:"from"`
	InvoicePayload  string           `json:"invoice_payload"`
	ShippingAddress *ShippingAddress `json:"shipping_address"`
}

// PreCheckoutQuery contains information about an incoming pre-checkout query
type PreCheckoutQuery struct {
	ID               string     `json:"id"`
	From             *User      `json:"from"`
	Currency         string     `json:"currency"`
	TotalAmount      int        `json:"total_amount"`
	InvoicePayload   string     `json:"invoice_payload"`
	ShippingOptionID string     `json:"shipping_option_id"`
	OrderInfo        *OrderInfo `json:"order_info"`
}

// Update represents an incoming update
// UpdateID is unique identifier
// At most one of the other fields can be not nil
type Update struct {
	UpdateID           int                 `json:"update_id"`
	Message            *Message            `json:"message,omitempty"`
	EditedMessage      *Message            `json:"edited_message,omitempty"`
	ChannelPost        *Message            `json:"channel_post,omitempty"`
	EditedChannelPost  *Message            `json:"edited_channel_pos,omitempty"`
	InlineQuery        *InlineQuery        `json:"inline_query,omitempty"`
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result,omitempty"`
	CallbackQuery      *CallbackQuery      `json:"callback_query,omitempty"`
	ShippingQuery      *ShippingQuery      `json:"shipping_query,omitempty"`
	PreCheckoutQuery   *PreCheckoutQuery   `json:"pre_checkout_query,omitempty"`
	Poll               *Poll               `json:"poll,omitempty"`
	PollAnswer         *PollAnswer         `json:"poll_answer,omitempty"`
}

// PassportData contains information about Telegram Passport data shared with the bot by the user
type PassportData struct {
	Data        []EncryptedPassportElement `json:"data"`
	Credentials EncryptedCredentials       `json:"credentials"`
}

// EncryptedPassportElement contains information about documents or other Telegram Passport elements shared with the bot by the user
type EncryptedPassportElement struct {
	Type        string          `json:"type"`
	Data        string          `json:"data"`
	PhoneNumber string          `json:"phone_number"`
	Email       string          `json:"email"`
	Files       []*PassportFile `json:"files"`
	FrontSide   *PassportFile   `json:"front_side"`
	ReverseSide *PassportFile   `json:"reverse_side"`
	Selfie      *PassportFile   `json:"selfie"`
}

// PassportFile represents a file uploaded to Telegram Passport
type PassportFile struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     int    `json:"file_size"`
	FileDate     int    `json:"file_date"`
}

// EncryptedCredentials contains data required for decrypting and authenticating EncryptedPassportElement
type EncryptedCredentials struct {
	Data   string `json:"data"`
	Hash   string `json:"hash"`
	Secret string `json:"secret"`
}

// Poll represents native telegram poll
type Poll struct {
	ID                    string       `json:"id"`
	Question              string       `json:"question"`
	Options               []PollOption `json:"options"`
	TotalVoterCount       int          `json:"total_voter_count"`
	IsClosed              bool         `json:"is_closed"`
	IsAnonymous           bool         `json:"is_anonymous"`
	Type                  string       `json:"type"`
	AllowsMultipleAnswers bool         `json:"allows_multiple_answers"`
	CorrectOptionID       int          `json:"correct_option_id"`
}

// Dice represents native telegram dice
type Dice struct {
	Emoji string `json:"emoji"`
	Value int    `json:"value"`
}

// PollOption is an option for Poll
type PollOption struct {
	Text       string `json:"text"`
	VoterCount int    `json:"voter_count"`
}

// PollAnswer represents an answer of a user in a non-anonymous poll
type PollAnswer struct {
	PollID    int   `json:"poll_id"`
	User      User  `json:"user"`
	OptionIDs []int `json:"option_ids"`
}

// LinkPreviewOptions https://core.telegram.org/bots/api#linkpreviewoptions
type LinkPreviewOptions struct {
	IsDisabled       bool   `json:"is_disabled,omitempty"`
	URL              string `json:"url,omitempty"`
	PreferSmallMedia bool   `json:"prefer_small_media,omitempty"`
	PreferLargeMedia bool   `json:"prefere_large_media,omitempty"`
	ShowAboveText    bool   `json:"show_above_text,omitempty"`
}

// ReplyParameters https://core.telegram.org/bots/api#replyparameters
type ReplyParameters struct {
	MessageID                int64           `json:"message_id"`
	ChatID                   int64           `json:"chat_id,omitempty"`
	AllowSendingWithoutReply bool            `json:"allow_sending_without_reply,omitempty"`
	Quote                    string          `json:"quote,omitempty"`
	QuoteParseMode           string          `json:"quote_parse_mode,omitempty"`
	QuoteEntities            []MessageEntity `json:"quote_entities,omitempty"`
	QuotePosition            int             `json:"quote_position,omitempty"`
}

// Set Stuff
type InputFile struct{}

type ParamSetWebhook struct {
	URL         *url.URL   `json:"url"`
	Certificate *InputFile `json:"certificate,omitempty"`
}

type ParamDeleteWebhook struct {
	DropPendingUpdates bool `json:"drop_pending_updates"`
}

// Sendble Types
// SendMesage https://core.telegram.org/bots/api#sendmessage
// SendPhoto
// SendVideo
// SendAnimation
// Send...
type Sendable struct {
	BusinessConnectionID string            `json:"business_connection_id,omitempty"`
	ChatID               int64             `json:"chat_id"`
	MessageThreadID      int64             `json:"message_thread_id,omitempty"`
	ParseMode            string            `json:"parse_mode,omitempty"`
	HasSpoiler           bool              `json:"has_spoiler,omitempty"`
	DisableNotification  bool              `json:"disable_notification,omitempty"`
	ProtectContent       bool              `json:"protect_content,omitempty"`
	AllowPaidBroadcast   bool              `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID      string            `json:"message_effect_id,omitempty"`
	ReplyParameters      []ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup          string            `json:"reply_markup,omitempty"`

	// Message
	Text               string              `json:"text,omitempty"`
	Entities           []MessageEntity     `json:"entities,omitempty"`
	LinkPreviewOptions *LinkPreviewOptions `json:"link_preview_options,omitempty"`

	// Attachments
	ShowCaptionAboveMedia bool            `json:"show_caption_above_media,omitempty"`
	Thumbnail             string          `json:"thumbnail,omitempty"`
	Caption               string          `json:"caption,omitempty"`
	CaptionEntities       []MessageEntity `json:"caption_entities,omitempty"`

	// SendPhoto https://core.telegram.org/bots/api#sendphoto
	Photo string `json:"photo,omitempty"`

	// SendAudio https://core.telegram.org/bots/api#sendaudio
	Audio string `json:"audio,omitempty"`

	// SendDocument https://core.telegram.org/bots/api#senddocument
	Document string `json:"document,omitempty"`

	// SendVideo https://core.telegram.org/bots/api#sendvideo
	Video             string `json:"video,omitempty"`
	Width             int    `json:"width,omitempty"`
	Height            int    `json:"heithg,omitempty"`
	Duration          int    `json:"duration,omitempty"`
	SupportsStreaming bool   `json:"supports_streaming,omitempty"`

	// SendMediaGroup https://core.telegram.org/bots/api#sendmediagroup
	Media []InputMedia `json:"media,omitempty"`

	// AnswerCallbackQuery
	CallbackQueryID string `json:"callback_query_id,omitempty"`
	ShowAlert       bool   `json:"show_alert,omitempty"`
	URL             string `json:"url,omitempty"`
	CacheTime       int    `json:"cache_time,omitempty"`

	// SetMessageReaction
	MessageID int64          `json:"message_id,omitempty"`
	Reaction  []ReactionType `json:"reaction_type,omitempty"`

	// Webhook
	Method string `json:"method,omitempty"`
}

// InputMedia
// InputMediaAudio https://core.telegram.org/bots/api#inputmediaaudio
// InputMediaVideo https://core.telegram.org/bots/api#inputmediavideo
// InputMediaDocument https://core.telegram.org/bots/api#inputmediadocument
type InputMedia struct {
	Type                  string          `json:"type"`
	Media                 string          `json:"media"`
	Thumbnail             string          `json:"thumbnail,omitempty"`
	Caption               string          `json:"caption,omitempty"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool            `json:"show_caption_above_media,omitempty"`
	HasSpoiler            bool            `json:"has_spoiler"`

	// InputMediaVideo
	Duration  int    `json:"duration,omitempty"`
	Performer string `json:"performer,omitempty"`
	Title     string `json:"title,omitempty"`

	// InputMediaDocument
	DisableContentTypeDetection bool `json:"disable_content_type_detection,omitempty"`
}

// InputPaidMedia https://core.telegram.org/bots/api#inputpaidmedia
// InputPaidMediaPhoto https://core.telegram.org/bots/api#inputpaidmediaphoto
type InputPaidMedia struct {
	Type  string `json:"type"`
	Media string `json:"media"`

	// InputPaidMediaVideo
	Thumbnail         string `json:"thumbnail,omitempty"`
	Width             int    `json:"width,omitempty"`
	Height            int    `json:"height,omitempty"`
	Duration          int    `json:"duration,omitempty"`
	SupportsStreaming bool   `json:"supports_streaming,omitempty"`
}

// ReactionType
// ReactionTypeEmoji
// ReactionTypeCustomEmoji
// ReactionTypePaid
type ReactionType struct {
	Type          string `json:"type"`
	Emoji         string `json:"emoji,omitempty"`
	CustomEmojiId string `json:"custom_emoji_id,omitempty"`
	TotalCount    int    `json:"total_count,omitempty"`
}

// BotCommand
type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type BotCommandScope struct {
	Type   string `json:"type"`
	ChatID int64  `json:"chat_id,omitempty"`
	UserID int64  `json:"user_id,omitempty"`
}

// SetMyCommandsParams
type SetMyCommandsParams struct {
	Commands     []BotCommand `json:"commands"`
	Scope        string       `json:"scope,omitempty"`
	LanguageCode string       `json:"language_code,omitempty"`
}
