package diagnostics

import (
	"fmt"
	"pascc/internal/models"
)

type DiagnosticInfo struct {
	ErrorCount   uint8
	WarningCount uint8
	InfoCount    uint8
}

func (di DiagnosticInfo) Any() bool {
	return di.ErrorCount > 0 || di.WarningCount > 0 || di.InfoCount > 0
}

func (di DiagnosticInfo) AnyError() bool {
	return di.ErrorCount > 0
}

type DiagnosticEmitter struct {
	diagnostics []Diagnostic
	DiagnosticInfo
}

func (d *DiagnosticEmitter) snapshot() DiagnosticEmitter {
	clone := *d
	return clone
}

func (d *DiagnosticEmitter) diagnosis(span models.Span, level DiagnosticLevel, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	Dbg("Filing diagnosis of type: ", level, ". span: ", span, ". with message: ", msg)
	d.diagnostics = append(d.diagnostics, Diagnostic{
		Span:  span,
		Msg:   msg,
		Level: level,
	})

	switch level {
	case (ERROR):
		d.DiagnosticInfo.ErrorCount++
	case (WARNING):
		d.DiagnosticInfo.WarningCount++
	case (INFO):
		d.DiagnosticInfo.InfoCount++
	}
}

func NewDiagnosticEmitter() DiagnosticEmitter {
	/* Create new Diagnostic Emitter */
	return DiagnosticEmitter{
		diagnostics: []Diagnostic{},
	}
}

func (d *DiagnosticEmitter) Diagnostics() []Diagnostic {
	return d.diagnostics
}

func (d *DiagnosticEmitter) Report(span models.Span, format string, args ...any) {
	d.diagnosis(span, ERROR, format, args...)
}

func (d *DiagnosticEmitter) Warn(span models.Span, format string, args ...any) {
	d.diagnosis(span, WARNING, format, args...)
}

func (d *DiagnosticEmitter) Info(span models.Span, format string, args ...any) {
	d.diagnosis(span, INFO, format, args...)
}

func (d *DiagnosticEmitter) Checkpoint() DiagnosticInfo {
	return d.DiagnosticInfo
}

func (d *DiagnosticEmitter) Diff(old DiagnosticInfo) DiagnosticInfo {
	return DiagnosticInfo{
		ErrorCount:   d.ErrorCount - old.ErrorCount,
		WarningCount: d.WarningCount - old.WarningCount,
		InfoCount:    d.InfoCount - old.InfoCount,
	}
}

func (d *DiagnosticEmitter) AnyErrorSince(checkpoint DiagnosticInfo) bool {
	return d.Diff(checkpoint).AnyError()
}
