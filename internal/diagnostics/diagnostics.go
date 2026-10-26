package diagnostics

import (
	"errors"
	"fmt"
	"pascc/internal/models"
	"strings"

	"github.com/fatih/color"
)

type DiagnosticLevel uint8

const (
	ERROR DiagnosticLevel = iota
	WARNING
	INFO
)

func (dl DiagnosticLevel) Color() color.Attribute {
	switch dl {
	case ERROR:
		return color.FgRed
	case WARNING:
		return color.FgYellow
	case INFO:
		return color.FgHiBlue
	default:
		return color.FgWhite
	}
}

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
	firstError := true

	for _, d := range ds {
		if d.Level != ERROR {
			continue
		}
		if !firstError {
			builder.WriteString(", ")
		}
		builder.WriteString(d.Msg)
		firstError = false
	}

	msg := builder.String()
	if len(msg) == 0 {
		return nil
	}

	return errors.New(builder.String())
}

/* Summarizes all diagnostics in the format (errors, warnings, info) */
func Summary(ds []Diagnostic) DiagnosticInfo {
	var e uint8 = 0
	var w uint8 = 0
	var i uint8 = 0

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

	return DiagnosticInfo{
		ErrorCount:   e,
		WarningCount: w,
		InfoCount:    i,
	}
}
