package telegram

import (
	"encoding/json"
)

type URLButton struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

type ReplyMarkup struct {
	InlineKeyboard string `json:"inline_keyboard"`
}

type WebhookResponse struct {
	Method string `json:"method,omitempty"`
	ChatID string `json:"chat_id"`

	// Media
	isMedia   bool
	Animation *string `json:"animation,omitempty"`
	Photo     *string `json:"photo,omitempty"`
	Video     *string `json:"video,omitempty"`
	Caption   *string `json:"caption,omitempty"`

	// Location
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	// Text
	Text      *string `json:"text,omitempty"`
	ParseMode string  `json:"parse_mode,omitempty"`

	// Keyboard
	ReplyMarkup *string `json:"reply_markup,omitempty"`
}

func (w *WebhookResponse) setMethod(method string, isMedia bool) *WebhookResponse {
	w.Method = method
	w.isMedia = isMedia
	return w
}

func (w *WebhookResponse) SendDice() *WebhookResponse {
	return w.setMethod("sendMessage", false).
		UseMarkdown()
}

func (w *WebhookResponse) SendMessage() *WebhookResponse {
	return w.setMethod("sendMessage", false)
}

func (w *WebhookResponse) SendVideo(id string) *WebhookResponse {
	w.Video = &id
	return w.setMethod("sendVideo", true)
}

func (w *WebhookResponse) SendAnimation(id string) *WebhookResponse {
	w.Animation = &id
	return w.setMethod("sendAnimation", true)
}

func (w *WebhookResponse) SendPhoto(id string) *WebhookResponse {
	w.Photo = &id
	return w.setMethod("sendPhoto", true)
}

func (w *WebhookResponse) SendLocation() *WebhookResponse {
	return w.setMethod("sendLocation", true)
}

func (w *WebhookResponse) SetChatID(chatID string) *WebhookResponse {
	w.ChatID = chatID
	return w
}

func (w *WebhookResponse) SetText(text string) *WebhookResponse {
	if w.isMedia {
		w.Caption = &text
	} else {
		w.Text = &text
	}
	return w
}

func (w *WebhookResponse) SetLinks(data []URLButton) *WebhookResponse {
	j, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	w.SetKeyboard(string(j))

	return w
}

func (w *WebhookResponse) UseMarkdown() *WebhookResponse {
	w.ParseMode = "MarkdownV2"
	return w
}

func (w *WebhookResponse) SetKeyboard(kb string) *WebhookResponse {
	w.ReplyMarkup = &kb
	return w
}

func (w *WebhookResponse) SetKeyboardInterface(kb interface{}) (*WebhookResponse, error) {
	data, err := json.Marshal(kb)
	if err != nil {
		return nil, err
	}
	dataString := string(data)
	w.ReplyMarkup = &dataString
	return w, nil
}

func (w *WebhookResponse) SetLocation(lat, lon float64) *WebhookResponse {
	w.Latitude = &lat
	w.Longitude = &lon

	return w
}

func (w *WebhookResponse) Empty() *WebhookResponse {
	w = &WebhookResponse{}
	return w
}

func (w *WebhookResponse) Marshal() ([]byte, error) {
	// empty response
	// spare 46 bytes
	if len(w.Method) == 0 && !w.isMedia {
		return []byte{}, nil
	}

	return json.Marshal(w)
}
