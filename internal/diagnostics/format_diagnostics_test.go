package diagnostics

import (
	"strings"
	"testing"

	"pascc/internal/models"
)

func TestFormatDiagnosticClampsColumnsPastLineEnd(t *testing.T) {
	// The lexer counts '\r' as a column, so a CRLF file can produce a span one
	// column past the visible end of a line.
	source := []byte("Algoritmo A\r\n\tEscribir 1\r\n\r\n")
	position := models.Position{Line: 3, Column: 2}
	out, err := FormatDiagnostic("a.psc", source, Diagnostic{
		Span:  models.Span{Start: position, End: position},
		Msg:   "expected keyword",
		Level: ERROR,
	})
	if err != nil {
		t.Fatalf("diagnostic could not be formatted: %v", err)
	}
	if !strings.Contains(out, "a.psc:3:1") {
		t.Fatalf("column was not clamped:\n%s", out)
	}
}
