package i18n

import "testing"

func TestCatalogsAreComplete(t *testing.T) {
	for lang, msgs := range catalog {
		for m := Msg(0); m < msgCount; m++ {
			if _, ok := msgs[m]; !ok {
				t.Errorf("language %d is missing message %d", lang, m)
			}
		}
	}
}
