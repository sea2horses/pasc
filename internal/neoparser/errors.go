package neoparser

import (
	"pseint-compiled/internal/lexer"
	"pseint-compiled/internal/models"
)

type ParseError struct {
	Position uint32

	Message  string
	Expected []Expectation
	Found    *lexer.Token

	Help   string
	Labels []Label
}

type ExpectationKind uint8

const (
	ExpectToken ExpectationKind = iota
	ExpectKeyword
	ExpectExpression
	ExpectType
	ExpectCustom
)

type Expectation struct {
	Kind ExpectationKind

	Token   lexer.TokenType
	Keyword lexer.Keyword
	Label   string
}

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

		found := ctx.Peek()

		return Match[T]{
			Kind:  Failed,
			Start: start,
			End:   ctx.Pos,
			Err: &ParseError{
				Position: ctx.Pos,
				Expected: []Expectation{expectation},
				Found:    found,
			},
		}
	})
}

func ExpectedToken(t lexer.TokenType) Expectation {
	return Expectation{
		Kind:  ExpectToken,
		Token: t,
	}
}

func ExpectedKeyword(k lexer.Keyword) Expectation {
	return Expectation{
		Kind:    ExpectKeyword,
		Keyword: k,
	}
}

func ExpectedCustom(label string) Expectation {
	return Expectation{
		Kind:  ExpectCustom,
		Label: label,
	}
}
