package parser

import (
	"os"
	"testing"

	"pascc/internal/i18n"
)

// Message assertions are written in English.
func TestMain(m *testing.M) {
	i18n.Set(i18n.EN)
	os.Exit(m.Run())
}
