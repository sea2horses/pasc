package lexer

import (
	"strings"
	"testing"
)

func TestUnterminatedStringStopsAtEndOfLine(t *testing.T) {
	source := "Escribir ' abre y cierra mal\"\nEscribir \"bien\"\n"
	lexer := NewLexer([]rune(source))
	tokens, _ := lexer.Tokenize()
	ds := lexer.Diagnostics()
	if len(ds) != 1 {
		t.Fatalf("expected one diagnostic, got %v", ds)
	}
	if ds[0].Span.Start.Line != 1 || !strings.Contains(ds[0].Msg, "different quote") {
		t.Fatalf("diagnostic should point at line 1 and mention the quote: %+v", ds[0])
	}
	found := false
	for _, token := range tokens {
		if token.Type == STRING_LITERAL && token.Value == "bien" && token.Span.Start.Line == 2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("lexing should resume on the next line: %v", tokens)
	}
}
