package diagnostics

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"pascc/internal/models"
)

func TestWriteJSONDiagnostics(t *testing.T) {
	ds := []Diagnostic{
		{
			Span: models.Span{
				Start: models.Position{Offset: 4, Line: 2, Column: 3},
				End:   models.Position{Offset: 7, Line: 2, Column: 6},
			},
			Msg:   `expected "<Entero>"`,
			Level: ERROR,
		},
		{Msg: "note", Level: INFO},
	}

	var out bytes.Buffer
	if err := WriteJSONDiagnostics(&out, ds, "a.psc"); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != len(ds) {
		t.Fatalf("got %d lines, want %d:\n%s", len(lines), len(ds), out.String())
	}

	want := `{"level":"error","message":"expected \"<Entero>\"","file":"a.psc","span":{"start":{"line":2,"column":3,"offset":4},"end":{"line":2,"column":6,"offset":7}}}`
	if lines[0] != want {
		t.Fatalf("line 1:\n got %s\nwant %s", lines[0], want)
	}

	var second JSONDiagnostic
	if err := json.Unmarshal([]byte(lines[1]), &second); err != nil {
		t.Fatal(err)
	}
	if second.Level != "info" || second.File != "a.psc" {
		t.Fatalf("line 2 decoded to %+v", second)
	}
}

func TestParseOutputFormat(t *testing.T) {
	for in, want := range map[string]OutputFormat{"": FormatText, "text": FormatText, "JSON": FormatJSON, " json ": FormatJSON} {
		got, err := ParseOutputFormat(in)
		if err != nil || got != want {
			t.Errorf("ParseOutputFormat(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	if _, err := ParseOutputFormat("xml"); err == nil {
		t.Error("ParseOutputFormat(\"xml\") should fail")
	}
}
