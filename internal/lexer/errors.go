package lexer

import (
	"pascc/internal/i18n"
)

type ErrOutOfBounds struct {
	Max   uint32
	Value uint32
}

func (e ErrOutOfBounds) Error() string {
	return i18n.T(i18n.CharOutOfBounds, e.Value, e.Max)
}

type ErrUnexpectedChar struct {
	Char rune
}

func (e ErrUnexpectedChar) Error() string {
	return i18n.T(i18n.UnexpectedChar, e.Char)
}

type LexerErr func() string

func (e LexerErr) Error() string {
	return e()
}

var (
	ErrExpectedIdentifier  = LexerErr(i18n.TFunc(i18n.ExpectedIdentifier))
	ErrExpectedString      = LexerErr(i18n.TFunc(i18n.ExpectedString))
	ErrExpectedNumber      = LexerErr(i18n.TFunc(i18n.ExpectedNumber))
	ErrUnterminatedString  = LexerErr(i18n.TFunc(i18n.UnterminatedString))
	ErrUnterminatedComment = LexerErr(i18n.TFunc(i18n.UnterminatedBlockComment))
)
