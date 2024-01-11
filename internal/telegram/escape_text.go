package telegram

import (
	"strings"
)

func escapeText(t string) string {
	t = strings.ReplaceAll(t, "*", `\*`)
	t = strings.ReplaceAll(t, "!", `\!`)
	t = strings.ReplaceAll(t, ".", `\.`)
	return t
}
