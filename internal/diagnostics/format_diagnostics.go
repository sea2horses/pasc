package diagnostics

import (
	"errors"
	"fmt"
	"io"
	"os"
	"pascc/internal/i18n"
	"pascc/internal/models"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
)

const (
	diagnosticContextLines = 1
	diagnosticTabWidth     = 4
)

// FormatDiagnostic creates a compiler-style diagnostic showing the selected
// span inside the source code.
//
// Example:
//
//	error: expected an integer
//	 --> example.lang:2:9
//	  |
//	1 | let x = 10
//	2 | let y = true
//	  |         ^^^^ expected an integer
//	3 | print(y)
func FormatDiagnostic(
	filename string,
	source []byte,
	diagnostic Diagnostic,
) (string, error) {
	normalizedSource := strings.ReplaceAll(string(source), "\r\n", "\n")
	lines := strings.Split(normalizedSource, "\n")
	diagnostic.Span = clampSpanColumns(lines, diagnostic.Span)

	if err := validateSpan(lines, diagnostic.Span); err != nil {
		return "", err
	}

	startLine := int(diagnostic.Span.Start.Line)
	endLine := int(diagnostic.Span.End.Line)

	firstVisibleLine := max(1, startLine-diagnosticContextLines)
	lastVisibleLine := min(len(lines), endLine+diagnosticContextLines)

	gutterWidth := len(strconv.Itoa(lastVisibleLine))

	var output strings.Builder

	fmt.Fprintf(&output, "%s: %s\n", diagnostic.Level, diagnostic.Msg)
	fmt.Fprintf(
		&output,
		" --> %s:%d:%d\n",
		filename,
		diagnostic.Span.Start.Line,
		diagnostic.Span.Start.Column,
	)
	fmt.Fprintf(&output, "%*s |\n", gutterWidth, "")

	messagePrinted := false

	for lineNumber := firstVisibleLine; lineNumber <= lastVisibleLine; lineNumber++ {
		line := lines[lineNumber-1]
		expandedLine := expandTabs(line, diagnosticTabWidth)

		fmt.Fprintf(
			&output,
			"%*d | %s\n",
			gutterWidth,
			lineNumber,
			expandedLine,
		)

		if lineNumber < startLine || lineNumber > endLine {
			continue
		}

		startColumn := uint32(1)
		endColumn := uint32(utf8.RuneCountInString(line) + 1)

		if lineNumber == startLine {
			startColumn = diagnostic.Span.Start.Column
		}

		if lineNumber == endLine {
			endColumn = diagnostic.Span.End.Column
		}

		// With an exclusive end position, a multiline span ending at column 1
		// does not include any characters from the final line.
		if startLine != endLine &&
			lineNumber == endLine &&
			endColumn == 1 {
			continue
		}

		startVisualColumn := visualColumn(
			line,
			startColumn,
			diagnosticTabWidth,
		)

		endVisualColumn := visualColumn(
			line,
			endColumn,
			diagnosticTabWidth,
		)

		markerLength := endVisualColumn - startVisualColumn
		if markerLength < 1 {
			// Show point diagnostics and empty spans with one caret.
			markerLength = 1
		}

		marker := strings.Repeat(" ", startVisualColumn) +
			strings.Repeat("^", markerLength)

		if !messagePrinted {
			marker += " " + diagnostic.Msg
			messagePrinted = true
		}

		fmt.Fprintf(
			&output,
			"%*s | %s\n",
			gutterWidth,
			"",
			marker,
		)
	}

	return output.String(), nil
}

// PrintFileDiagnostic reads a file and prints its diagnostic to stderr.
func PrintFileDiagnostic(
	filename string,
	diagnostic Diagnostic,
) error {
	source, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T(i18n.ReadSourceFailed), err)
	}

	f, err := FormatDiagnostic(filename, source, diagnostic)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T(i18n.FormatDiagnosticFailed), err)
	}

	// Set color
	color.Set(diagnostic.Level.Color())
	defer color.Unset()

	_, err = io.WriteString(os.Stderr, f)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T(i18n.WriteDiagnosticFailed), err)
	}

	return nil
}

func PrintDiagnostics(ds []Diagnostic, filename string) {
	if Format == FormatJSON {
		if err := WriteJSONDiagnostics(os.Stderr, ds, filename); err != nil {
			fmt.Fprintln(os.Stderr, i18n.T(i18n.WriteDiagnosticFailed), err)
		}
		return
	}
	for i, d := range ds {
		err := PrintFileDiagnostic(filename, d)
		if err != nil {
			fmt.Fprintln(os.Stderr, i18n.T(i18n.DiagnosticNotShown, i+1, err.Error()))
		}
	}
	PrintSummary(Summary(ds))
}

func PrintSummary(d DiagnosticInfo) {
	parts := []struct {
		count     uint8
		one, many i18n.Msg
		level     DiagnosticLevel
	}{
		{d.ErrorCount, i18n.SummaryError, i18n.SummaryErrors, ERROR},
		{d.WarningCount, i18n.SummaryWarning, i18n.SummaryWarnings, WARNING},
		{d.InfoCount, i18n.SummaryNote, i18n.SummaryNotes, INFO},
	}
	printed := false
	for _, part := range parts {
		if part.count == 0 {
			continue
		}
		if printed {
			fmt.Print(" | ")
		}
		msg := part.many
		if part.count == 1 {
			msg = part.one
		}
		color.Set(part.level.Color())
		fmt.Print(i18n.T(msg, part.count))
		color.Unset()
		printed = true
	}
	if printed {
		fmt.Println()
	}
}

func validateSpan(lines []string, span models.Span) error {
	if span.Start.Line == 0 || span.Start.Column == 0 {
		return errors.New(i18n.T(i18n.SpanStartNotOneBased, span.Start.Line, span.Start.Column))
	}

	if span.End.Line == 0 || span.End.Column == 0 {
		return errors.New(i18n.T(i18n.SpanEndNotOneBased, span.End.Line, span.End.Column))
	}

	if span.Start.Line > span.End.Line ||
		(span.Start.Line == span.End.Line &&
			span.Start.Column > span.End.Column) {
		return errors.New(i18n.T(i18n.SpanStartAfterEnd))
	}

	if int(span.Start.Line) > len(lines) {
		return errors.New(i18n.T(i18n.SpanStartOutsideFile, span.Start.Line))
	}

	if int(span.End.Line) > len(lines) {
		return errors.New(i18n.T(i18n.SpanEndOutsideFile, span.End.Line))
	}

	if err := validatePositionColumn(lines, span.Start, "start"); err != nil {
		return err
	}

	if err := validatePositionColumn(lines, span.End, "end"); err != nil {
		return err
	}

	return nil
}

func validatePositionColumn(
	lines []string,
	position models.Position,
	name string,
) error {
	line := lines[position.Line-1]
	maxColumn := uint32(utf8.RuneCountInString(line) + 1)

	if position.Column > maxColumn {
		return errors.New(i18n.T(i18n.SpanColumnTooLarge, position.Column, position.Line, maxColumn))
	}

	return nil
}

// visualColumn converts a 1-based source column into a zero-based visual
// column, accounting for tab expansion.
func visualColumn(line string, column uint32, tabWidth int) int {
	targetRune := int(column) - 1
	currentRune := 0
	currentVisualColumn := 0

	for _, character := range line {
		if currentRune >= targetRune {
			break
		}

		if character == '\t' {
			currentVisualColumn += tabWidth -
				(currentVisualColumn % tabWidth)
		} else {
			currentVisualColumn++
		}

		currentRune++
	}

	return currentVisualColumn
}

func expandTabs(line string, tabWidth int) string {
	var expanded strings.Builder
	visualColumn := 0

	for _, character := range line {
		if character == '\t' {
			spaceCount := tabWidth - (visualColumn % tabWidth)
			expanded.WriteString(strings.Repeat(" ", spaceCount))
			visualColumn += spaceCount
			continue
		}

		expanded.WriteRune(character)
		visualColumn++
	}

	return expanded.String()
}

// clampSpanColumns keeps columns within their line. The lexer counts '\r' as
// a column while the formatter strips it, so spans near the end of a line in
// a CRLF file can point one column past the visible text.
func clampSpanColumns(lines []string, span models.Span) models.Span {
	clamp := func(position models.Position) models.Position {
		if position.Line == 0 || int(position.Line) > len(lines) {
			return position
		}
		maxColumn := uint32(utf8.RuneCountInString(lines[position.Line-1]) + 1)
		position.Column = min(position.Column, maxColumn)
		return position
	}
	span.Start = clamp(span.Start)
	span.End = clamp(span.End)
	return span
}
