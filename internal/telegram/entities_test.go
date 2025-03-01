package telegram

import "testing"

func TestUTF16Len(t *testing.T) {
	testCases := map[string]struct {
		i string
		l int
	}{
		"empty": {"", 0},
	}
	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			res := UTF16Len(test.i)
			if res != test.l {
				t.Errorf("expeted len %d, returned len %d", res, test.l)
			}
		})
	}
}
