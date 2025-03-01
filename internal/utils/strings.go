package utils

import "strings"

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\n' || r == 0
}

func FirstWord(s string) string {
	if len(s) == 0 {
		return s
	}

	b := strings.Builder{}

	for _, r := range s {
		if isWhitespace(r) {
			break
		}
		b.WriteRune(r)
	}

	return b.String()
}

func SplitMessageWords(s string) (string, string) {
	if len(s) == 0 {
		return s, s
	}

	head := strings.Builder{}
	rest := strings.Builder{}
	for idx, r := range s {
		// build rest when encountering a whitespace or EOF
		if isWhitespace(r) {
			rest.WriteString(s[idx+1:])
			break
		}

		if _, err := head.WriteRune(r); err != nil {
			panic(err)
		}
	}
	rs := strings.Trim(rest.String(), " ")
	hs := strings.Trim(head.String(), " ")

	return hs, rs
}
