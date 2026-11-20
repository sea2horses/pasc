package analyzer

import (
	"pascc/internal/i18n"
	"pascc/internal/models"
	"pascc/internal/semantic"
)

type ErrIncorrectType struct {
	Expected semantic.Type
	Got      semantic.Type
}

func (e ErrIncorrectType) Error() string {
	return i18n.T(i18n.IncorrectType, semantic.DisplayType(e.Expected), semantic.DisplayType(e.Got))
}

type ErrUnOperationNotAllowed struct {
	Operation semantic.OperatorType
	Type      semantic.Type
}

func (e ErrUnOperationNotAllowed) Error() string {
	return i18n.T(i18n.UnaryOpNotAllowed, e.Operation.Display(), semantic.DisplayType(e.Type))
}

type ErrBiOperationNotAllowed struct {
	Operation semantic.OperatorType
	LHS       semantic.Type
	RHS       semantic.Type
}

func (e ErrBiOperationNotAllowed) Error() string {
	return i18n.T(i18n.BinaryOpNotAllowed, e.Operation.Display(), semantic.DisplayType(e.LHS), semantic.DisplayType(e.RHS))
}

type ErrVariableDoesntExist struct {
	Name string
}

func (e ErrVariableDoesntExist) Error() string {
	return i18n.T(i18n.VariableDoesntExist, e.Name)
}

type ErrSymbolAlreadyExists struct {
	Name string
}

func (e ErrSymbolAlreadyExists) Error() string {
	return i18n.T(i18n.SymbolAlreadyExists, e.Name)
}

// ErrUnsupportedType is a value that Escribir cannot display.
type ErrUnsupportedType struct {
	Type semantic.Type
}

func (e ErrUnsupportedType) Error() string {
	return i18n.T(i18n.UnsupportedWriteType, semantic.DisplayType(e.Type))
}

type ErrExpectedType struct {
	Type semantic.Type
}

func (e ErrExpectedType) Error() string {
	return i18n.T(i18n.ExpectedValueOfType, semantic.DisplayType(e.Type))
}

// AnalyzerError is an argument-less error backed by a catalog message.
type AnalyzerError i18n.Msg

func (e AnalyzerError) Error() string { return i18n.T(i18n.Msg(e)) }

const (
	ErrUnexpectedVoid = AnalyzerError(i18n.UnexpectedVoid)
	ErrLValueError    = AnalyzerError(i18n.NotAssignable)
	ErrImmutable      = AnalyzerError(i18n.Immutable)
)

// errorf, warnf and infof report a translated catalog message.
func (a *Analyzer) errorf(span models.Span, m i18n.Msg, args ...any) {
	a.Report(span, "%s", i18n.T(m, args...))
}

func (a *Analyzer) warnf(span models.Span, m i18n.Msg, args ...any) {
	a.Warn(span, "%s", i18n.T(m, args...))
}

func (a *Analyzer) infof(span models.Span, m i18n.Msg, args ...any) {
	a.Info(span, "%s", i18n.T(m, args...))
}

func (a *Analyzer) reportErr(span models.Span, err error) {
	a.Report(span, "%s", err.Error())
}
