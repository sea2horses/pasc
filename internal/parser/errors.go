package parser

import (
	"pascc/internal/i18n"
	"pascc/internal/lexer"
)

// ParseError records an absolute token position at the point of failure.
// The position must survive cursor rollback and recovery; len(Tokens) denotes EOF.
type ParseError interface {
	Error() string
	TokenPosition() uint32
}

// Expectation describes what the parser wanted at a failure point.
// Expected is rendered lazily so the active language is applied at report time.
type Expectation interface {
	Expected() string
	Error() string
}

// LiteralExpectations are argument-less expectations backed by a catalog message.
type LiteralExpectations i18n.Msg

const (
	ExpectedContinuation = LiteralExpectations(i18n.ExpectedContinuation)
	ExpectedExpression   = LiteralExpectations(i18n.ExpectedExpression)
	ExpectedNewline      = LiteralExpectations(i18n.ExpectedNewline)
	ExpectedType         = LiteralExpectations(i18n.ExpectedType)
	ExpectedStatement    = LiteralExpectations(i18n.ExpectedStatement)
	ExpectedSwitchCase   = LiteralExpectations(i18n.ExpectedSwitchCase)
	ExpectedTimeUnit     = LiteralExpectations(i18n.ExpectedTimeUnit)
	ExpectedEndOfProgram = LiteralExpectations(i18n.ExpectedEndOfProgram)
	InvalidSyntax        = LiteralExpectations(i18n.InvalidSyntax)

	ExpectedDelimitedContent = LiteralExpectations(i18n.ExpectedDelimitedContent)
	ExpectedClosingDelimiter = LiteralExpectations(i18n.ExpectedClosingDelimiter)
)

func (e LiteralExpectations) Expected() string { return i18n.T(i18n.Msg(e)) }
func (e LiteralExpectations) Error() string    { return e.Expected() }

type ExpectedToken struct{ Kind lexer.TokenType }

func (e ExpectedToken) Expected() string {
	switch e.Kind {
	case lexer.IDENTIFIER:
		return i18n.T(i18n.ExpectedIdentifier)
	case lexer.NUMBER_LITERAL:
		return i18n.T(i18n.ExpectedNumber)
	case lexer.STRING_LITERAL:
		return i18n.T(i18n.ExpectedString)
	case lexer.NEWLINE:
		return ExpectedNewline.Expected()
	}
	if symbol, ok := e.Kind.Symbol(); ok {
		return i18n.T(i18n.ExpectedSymbol, symbol)
	}
	return i18n.T(i18n.ExpectedTokenKind, e.Kind)
}
func (e ExpectedToken) Error() string { return e.Expected() }

type ExpectedKeyword struct{ Keyword lexer.Keyword }

func (e ExpectedKeyword) Expected() string {
	return i18n.T(i18n.ExpectedKeyword, e.Keyword.Display())
}
func (e ExpectedKeyword) Error() string { return e.Expected() }

// ExpectedClosingKeyword is a missing block terminator, e.g. FinSi for Si.
type ExpectedClosingKeyword struct{ Close, Open lexer.Keyword }

func (e ExpectedClosingKeyword) Expected() string {
	return i18n.T(i18n.ExpectedClosingKeyword, e.Close.Display(), e.Open.Display())
}
func (e ExpectedClosingKeyword) Error() string { return e.Expected() }

// ExpectedItemAfterSeparator is a dangling separator, e.g. `foo(1,)`.
type ExpectedItemAfterSeparator struct{ Separator string }

func (e ExpectedItemAfterSeparator) Expected() string {
	return i18n.T(i18n.ExpectedItemAfterSeparator, e.Separator)
}
func (e ExpectedItemAfterSeparator) Error() string { return e.Expected() }

// SyntaxError keeps an absolute token position even when the cursor is rolled back.
type SyntaxError struct {
	Position uint32
	Expected Expectation
	Found    *lexer.Token
}

func (e *SyntaxError) TokenPosition() uint32 { return e.Position }

func (e *SyntaxError) Error() string {
	if e.Found == nil || e.Found.Type == lexer.EOF {
		return i18n.T(i18n.SyntaxErrorAtEOF, e.Expected.Expected())
	}
	return i18n.T(i18n.SyntaxErrorFound, e.Expected.Expected(), describeToken(e.Found))
}

// describeToken phrases a token the way the user wrote it.
func describeToken(t *lexer.Token) string {
	switch t.Type {
	case lexer.NEWLINE:
		return i18n.T(i18n.FoundNewline)
	case lexer.KEYWORD:
		return i18n.T(i18n.FoundKeyword, t.Value)
	case lexer.STRING_LITERAL:
		return i18n.T(i18n.FoundString, t.Value)
	}
	return i18n.T(i18n.FoundToken, t.Value)
}

func failureAt(ctx *Context, pos uint32, expected Expectation) *SyntaxError {
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
		return Match[T]{Kind: Failed, Start: start, End: pos, Err: failureAt(ctx, pos, expectation)}
	})
}
