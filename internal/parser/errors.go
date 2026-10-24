package parser

import (
	"fmt"
	"pascc/internal/lexer"
)

type ParserError string

func (pe ParserError) Error() string {
	return string(pe)
}

const (
	ErrOutOfBounds         = ParserError("index is out of bounds")
	ErrUnexpectedEOF       = ParserError("unexpected EOF (end of file)")
	ErrUnrecognizedKeyword = ParserError("unrecognized keyword")
	ErrNotImplemented      = ParserError("not implemented")
	ErrExpectedOperand     = ParserError("expected operand")
	ErrUnexpectedNode      = ParserError("unexpected node in expression")
	ErrInvalidExpression   = ParserError("invalid expression")
	ErrExpectedType        = ParserError("expected type")
	ErrInvalidStatement    = ParserError("statement contains errors")
)

type ErrExpectedToken struct {
	Got      lexer.TokenType
	Expected lexer.TokenType
}

func (er ErrExpectedToken) Error() string {
	return fmt.Sprintf("expected token type: %s, got: %s", er.Expected, er.Got)
}

type ErrExpectedKeyword struct {
	Expected lexer.Keyword
}

func (er ErrExpectedKeyword) Error() string {
	return fmt.Sprintf("expected keyword: %s", er.Expected)
}

type ErrUnterminatedDirective struct {
	Name string
}

func (ud ErrUnterminatedDirective) Error() string {
	return fmt.Sprintf("unterminated directive: %s", ud.Name)
}
