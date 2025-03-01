package utils

import (
	"testing"
)

func TestSplitMessageWords(t *testing.T) {
	testSamples := map[string]struct {
		head int
		rest int
		text string
	}{
		"simple":       {4, 19, "this is a simple message"},
		"empty":        {0, 0, ""},
		"command":      {4, 0, "/foo"},
		"command rest": {4, 9, "/bar rest rest"},
	}

	for name, sample := range testSamples {
		t.Run(name, func(t *testing.T) {
			head, rest := SplitMessageWords(sample.text)
			if len(head) != sample.head {
				t.Errorf(
					"head '%s' should have len %d, found: %d",
					head, sample.head, len(head),
				)
			}
			if len(rest) != sample.rest {
				t.Errorf(
					"rest '%s' should have len %d, found: %d",
					rest, sample.rest, len(rest),
				)
			}
		})
	}
}
