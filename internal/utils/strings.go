package utils

import (
	"strings"
	"unicode"
)

func FirstWord(s string) string {
	if len(s) == 0 {
		return s
	}

	b := strings.Builder{}

	for _, r := range s {
		if unicode.IsSpace(r) {
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
		if unicode.IsSpace(r) {
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

func CleanText(s string) string {
	b := strings.Builder{}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
