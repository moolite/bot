package tg

import (
	"encoding/json"
)

func UnmarshalUpdate(data []byte) (*Update, error) {
	u := &Update{}
	if err := json.Unmarshal(data, u); err != nil {
		return u, err
	}
	return u, nil
}

// UTF16Len length of sub-strings using telegram's UTF16 entity length function
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
