package neoparser

import (
	"pseint-compiled/internal/lexer"
	"slices"
)

type RecoveryStrategy interface {
	Recover(*Context) RecoveryResult
}

type RecoveryResult struct {
	Start uint32
	End   uint32
}

// Recover and produce a fallback AST value.
func RecoverWith[T any](
	p Pattern[T],
	strategy RecoveryStrategy,
	fallback func(*ParseError) T,
) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		result := p.Match(ctx)

		if result.Kind == Failed {
			strategy.Recover(ctx)
			return Match[T]{
				Kind:  Matched,
				Value: fallback(result.Err),
				Start: result.Start,
				End:   result.End,
				Err:   result.Err,
			}
		}

		return result
	})
}

func Recover[T any](
	p Pattern[T],
	strategy RecoveryStrategy,
) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		result := p.Match(ctx)

		if result.Kind == Failed {
			strategy.Recover(ctx)
		}

		return result
	})
}

type Synchronize struct {
	Tokens   []lexer.TokenType
	Keywords []lexer.Keyword
}

func (s *Synchronize) Recover(ctx *Context) {
	for !ctx.EOF() {
		tok := ctx.Peek()

		kw, ok := lexer.MapToKeyword([]rune(tok.Value))
		if ok && slices.Contains(s.Keywords, kw) || slices.Contains(s.Tokens, tok.Type) {
		}
		ctx.Advance()
	}
}

type UntilNewline struct{}

func (UntilNewline) Recover(ctx *Context) {
	for !ctx.EOF() {
		if ctx.Peek().Type == lexer.NEWLINE {
			return
		}

		ctx.Advance()
	}
}

type UntilKeywords struct {
	Keywords []lexer.Keyword
}

func (uk UntilKeywords) Recover(ctx *Context) {
	for !ctx.EOF() {
		tok := ctx.Peek()

		kw, ok := lexer.MapToKeyword([]rune(tok.Value))
		if tok.Type == lexer.KEYWORD && ok && slices.Contains(uk.Keywords, kw) {
			return
		}

		ctx.Advance()
	}
}
