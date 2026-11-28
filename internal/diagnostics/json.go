package diagnostics

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"pascc/internal/i18n"
	"pascc/internal/models"
)

type OutputFormat uint8

const (
	FormatText OutputFormat = iota /* Human readable, with source excerpts and colors */
	FormatJSON                     /* One JSON object per line (JSON Lines) on stderr */
)

// Format selects how diagnostics are emitted. Like Debug, it is set once by
// the command line before compiling.
var Format = FormatText

func ParseOutputFormat(value string) (OutputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "text", "human":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	}
	return FormatText, errors.New(i18n.T(i18n.InvalidDiagnosticFormat, value))
}

// Key returns the level's stable, untranslated name for machine output.
func (dl DiagnosticLevel) Key() string {
	switch dl {
	case ERROR:
		return "error"
	case WARNING:
		return "warning"
	case INFO:
		return "info"
	}
	return "unknown"
}

type jsonPosition struct {
	Line   uint32 `json:"line"`   /* 1-based */
	Column uint32 `json:"column"` /* 1-based, counted in Unicode code points */
	Offset uint32 `json:"offset"` /* 0-based, counted in Unicode code points */
}

type jsonSpan struct {
	Start jsonPosition `json:"start"`
	End   jsonPosition `json:"end"` /* exclusive */
}

// JSONDiagnostic is the shape of each line written in FormatJSON. Fatal
// errors that do not come from the source (missing file, failed build...)
// have no file or span.
type JSONDiagnostic struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	File    string    `json:"file,omitempty"`
	Span    *jsonSpan `json:"span,omitempty"`
}

func toJSONPosition(p models.Position) jsonPosition {
	return jsonPosition{Line: p.Line, Column: p.Column, Offset: p.Offset}
}

func ToJSON(d Diagnostic, filename string) JSONDiagnostic {
	return JSONDiagnostic{
		Level:   d.Level.Key(),
		Message: d.Msg,
		File:    filename,
		Span: &jsonSpan{
			Start: toJSONPosition(d.Span.Start),
			End:   toJSONPosition(d.Span.End),
		},
	}
}

func writeJSON(w io.Writer, v JSONDiagnostic) error {
	/* Encoder.Encode terminates every value with a newline */
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func WriteJSONDiagnostics(w io.Writer, ds []Diagnostic, filename string) error {
	for _, d := range ds {
		if err := writeJSON(w, ToJSON(d, filename)); err != nil {
			return err
		}
	}
	return nil
}

// PrintError reports an error that ends the program, in the current Format.
func PrintError(err error) {
	if Format == FormatJSON {
		if writeJSON(os.Stderr, JSONDiagnostic{Level: ERROR.Key(), Message: err.Error()}) == nil {
			return
		}
	}
	io.WriteString(os.Stderr, err.Error()+"\n")
}
