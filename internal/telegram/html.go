package telegram

import "fmt"

func htmlEl(s, el string) string {
	return fmt.Sprintf("<%s>%s</%s>", el, s, el)
}

func HtmlB(s string) string {
	return htmlEl(s, "b")
}

func HtmlI(s string) string {
	return htmlEl(s, "i")
}

func HtmlCode(s string) string {
	return htmlEl(s, "code")
}

func HtmlS(s string) string {
	return htmlEl(s, "s")
}

func HtmlU(s string) string {
	return htmlEl(s, "u")
}

func HtmlPre(s, language string) string {
	return fmt.Sprintf("<pre language=\"%s\">%s</pre>", language, s)
}
