package telegram

import (
	"strings"

	"github.com/go-telegram/bot/models"
)

// telegram docs: https://core.telegram.org/bots/api#messageentity
func AsHTML(text string, entities []models.MessageEntity) string {
	if len(entities) == 0 {
		return text
	}

	r := strings.Builder{}
	for _, entity := range entities {
		switch entity.Type {
		case "mention":
		case "hashtag":
		case "cashtag":
		case "bot_command":
		case "url":
		case "email":
		case "phone_number":
		case "bold":
		case "italic":
		case "underline":
		case "strikethrough":
			a := text[entity.Offset : entity.Offset+entity.Length]
			r.WriteString("<b>")
			r.WriteString(a)
			r.WriteString("</b>")
		case "spoiler":
		case "blockquote":
		case "expandable_blockquote":
		case "code":
		case "pre":
		case "text_link":
		case "text_mention":
		case "custom_emoji":
		}

	}

	return r.String()
}

func UTF16Len(s string) int {
	l := 0
	for _, r := range s {
		if r&0xc0 != 0x80 {
			if r >= 0xf0 {
				l += 2
			} else {
				l += 1
			}
		}
	}

	return l
}
