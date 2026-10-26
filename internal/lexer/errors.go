package lexer

import "fmt"

type ErrOutOfBounds struct {
	Max   uint32
	Value uint32
}

func (e ErrOutOfBounds) Error() string {
	return fmt.Sprintf("value %d is out of bounds (%d)", e.Value, e.Max)
}

type ErrUnexpectedChar struct {
	Char rune
}

func (e ErrUnexpectedChar) Error() string {
	return fmt.Sprintf("unexpected char: %c", e.Char)
}

type LexerErr string

func (e LexerErr) Error() string {
	return string(e)
}

const (
	ErrExpectedIdentifier  = LexerErr("expected identifier")
	ErrExpectedString      = LexerErr("expected string")
	ErrExpectedNumber      = LexerErr("expected number")
	ErrUnterminatedString  = LexerErr("unterminated string")
	ErrUnterminatedComment = LexerErr("unterminated block comment")
)
