package neoparser

import (
	"fmt"
	"pseint-compiled/internal/lexer"
)

type ParseError interface{ Error() string }
type LiteralErrors string

const ItemAfterSeparator LiteralErrors = "expected item after separator"

func (e LiteralErrors) Error() string { return string(e) }

type Expectation interface {
	Expected() string
	Error() string
}
type LiteralExpectations string

const (
	ExpectedContinuation LiteralExpectations = "expected continuation"
	ExpectedExpression   LiteralExpectations = "expected expression"
	ExpectedNewline      LiteralExpectations = "expected newline"
)

func (e LiteralExpectations) Expected() string { return string(e) }
func (e LiteralExpectations) Error() string    { return e.Expected() }

type ExpectedToken struct{ Kind lexer.TokenType }

func (e ExpectedToken) Expected() string { return fmt.Sprintf("expected token '%s'", e.Kind) }
func (e ExpectedToken) Error() string    { return e.Expected() }

type ExpectedKeyword struct{ Keyword lexer.Keyword }

func (e ExpectedKeyword) Expected() string { return fmt.Sprintf("expected keyword '%s'", e.Keyword) }
func (e ExpectedKeyword) Error() string    { return e.Expected() }

// SyntaxError keeps an absolute token position even when the cursor is rolled back.
type SyntaxError struct {
	Position uint32
	Expected string
	Found    *lexer.Token
}

func (e *SyntaxError) Error() string {
	if e.Found == nil {
		return fmt.Sprintf("%s at end of input", e.Expected)
	}
	return fmt.Sprintf("%s, found %q", e.Expected, e.Found.Value)
}
func failureAt(ctx *Context, pos uint32, expected string) *SyntaxError {
	var found *lexer.Token
	if pos < uint32(len(ctx.Tokens)) {
		found = &ctx.Tokens[pos]
	}
	return &SyntaxError{Position: pos, Expected: expected, Found: found}
}
func Expected[T any](p Pattern[T], expectation Expectation) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		start := ctx.Pos
		result := p.Match(ctx)
		if result.Kind != NoMatch {
			return result
		}
		pos := max(ctx.Pos, result.End)
		return Match[T]{Kind: Failed, Start: start, End: pos, Err: failureAt(ctx, pos, expectation.Expected())}
	})
}
