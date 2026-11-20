package diagnostics

import (
	"errors"
	"fmt"
	"pascc/internal/i18n"
	"pascc/internal/models"

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
		return i18n.T(i18n.LevelError)
	case WARNING:
		return i18n.T(i18n.LevelWarning)
	case INFO:
		return i18n.T(i18n.LevelInfo)
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

// SummaryAsError returns nil when ds has no errors. The diagnostics themselves
// are printed separately, so the error only states that compilation failed.
func SummaryAsError(ds []Diagnostic) error {
	count := Summary(ds).ErrorCount
	switch count {
	case 0:
		return nil
	case 1:
		return errors.New(i18n.T(i18n.CompilationFailed, count))
	}
	return errors.New(i18n.T(i18n.CompilationFailedMany, count))
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
