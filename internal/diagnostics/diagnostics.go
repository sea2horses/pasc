package diagnostics

import (
	"errors"
	"fmt"
	"pascc/internal/models"
	"strings"
)

type DiagnosticLevel uint8

const (
	ERROR DiagnosticLevel = iota
	WARNING
	INFO
)

func (dl DiagnosticLevel) String() string {
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

type Diagnostic struct {
	Span  models.Span
	Msg   string
	Level DiagnosticLevel
}

type Diagnostics struct {
	Src  []rune
	list []Diagnostic
}

func (d *Diagnostics) Errorf(s models.Span, level DiagnosticLevel, f string, a ...any) {
	d.list = append(d.list, Diagnostic{
		Span:  s,
		Msg:   fmt.Sprintf(f, a...),
		Level: level,
	})
}

func SummaryAsError(ds []Diagnostic) error {
	var builder strings.Builder

	for i, d := range ds {
		if i > 0 {
			builder.WriteString(", ")
		}

		if d.Level == ERROR {
			builder.WriteString(d.Msg)
		}
	}

	return errors.New(builder.String())
}

/* Summarizes all diagnostics in the format (errors, warnings, info) */
func Summary(ds []Diagnostic) (int, int, int) {
	e := 0
	w := 0
	i := 0

	for _, d := range ds {
		switch d.Level {
		case ERROR:
			e++
		case WARNING:
			w++
		case INFO:
			i++
		}
	}

	return e, w, i
}
