package neoparser

import (
	"fmt"
	"pseint-compiled/internal/lexer"
)

type ParseError interface {
	Error() string
}

type LiteralErrors string

const (
	ItemAfterSeparator LiteralErrors = "expected item after separator"
)

func (le LiteralErrors) Error() string {
	return string(le)
}

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

func (le LiteralExpectations) Expected() string {
	return string(le)
}

func (le LiteralExpectations) Error() string { return le.Expected() }

type ExpectedToken struct {
	Kind lexer.TokenType
}

func (et ExpectedToken) Expected() string {
	return fmt.Sprintf("expected token '%s'", et.Kind)
}

func (et ExpectedToken) Error() string { return et.Expected() }

type ExpectedKeyword struct {
	Keyword lexer.Keyword
}

func (kw ExpectedKeyword) Expected() string {
	return fmt.Sprintf("expected keyword '%s'", kw.Keyword)
}

func (kw ExpectedKeyword) Error() string { return kw.Expected() }

func Expected[T any](
	p Pattern[T],
	expectation Expectation,
) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		start := ctx.Pos

		result := p.Match(ctx)

		if result.Kind != NoMatch {
			return result
		}

		return Match[T]{
			Kind:  Failed,
			Start: start,
			End:   ctx.Pos,
			Err:   expectation,
		}
	})
}
