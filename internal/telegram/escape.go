package telegram

import (
	"strings"
)

// /([_*[\]()~`>#+\-=|{}.!])/g
var mkdCharacters = [...]string{
	`.`, `_`, `\`, `*`, `[`, `]`, `(`, `)`,
	`~`, "`", `<`, `>`, `#`, `+`, `-`, `=`,
	`|`, `{`, `}`, `!`,
}

func EscapeNonMarkdownText(text string) string {
	return strings.ReplaceAll(text, `.`, `\.`)
}
