package neoparser

import (
	"fmt"
	"pseint-compiled/internal/lexer"
	"pseint-compiled/internal/models"
)

type ParseError interface {
	Error() string
}

type LiteralErrors string

const (
	ItemAfterSeparator LiteralErrors = "expected item after separator"
)

func (le LiteralErrors) Error() string {
	return fmt.Sprintf("%s", le)
}

type Expectation interface {
	Expected() string
	Error() string
}

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

type ExpectedExpression struct{}

func (ee ExpectedExpression) Expected() string { return "expected expression" }
func (ee ExpectedExpression) Error() string    { return ee.Expected() }

type ExpectedNewline struct{}

func (ee ExpectedNewline) Expected() string { return "expected newline" }
func (ee ExpectedNewline) Error() string    { return ee.Expected() }

type Label struct {
	Span    models.Span
	Message string
}

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
