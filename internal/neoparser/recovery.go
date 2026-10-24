package neoparser

import (
	"pseint-compiled/internal/lexer"
	"slices"
)

type RecoveryStrategy interface{ Recover(*Context) RecoveryResult }
type RecoveryResult struct{ Start, End uint32 }

// Synchronization stops BEFORE the boundary, leaving it for the enclosing parser.
type Synchronize struct {
	Tokens   []lexer.TokenType
	Keywords []lexer.Keyword
}

func (s Synchronize) Recover(ctx *Context) RecoveryResult {
	start := ctx.Pos
	for !ctx.EOF() {
		tok := ctx.Peek()
		if slices.Contains(s.Tokens, tok.Type) {
			break
		}
		if tok.Type == lexer.KEYWORD {
			if kw, ok := lexer.MapToKeyword([]rune(tok.Value)); ok && slices.Contains(s.Keywords, kw) {
				break
			}
		}
		ctx.Advance()
	}
	return RecoveryResult{Start: start, End: ctx.Pos}
}

type UntilNewline struct{}

func (UntilNewline) Recover(ctx *Context) RecoveryResult {
	return Synchronize{Tokens: []lexer.TokenType{lexer.NEWLINE}}.Recover(ctx)
}

type UntilKeywords struct{ Keywords []lexer.Keyword }

func (s UntilKeywords) Recover(ctx *Context) RecoveryResult {
	return Synchronize{Keywords: s.Keywords}.Recover(ctx)
}
func RecoverWith[T any](p Pattern[T], strategy RecoveryStrategy, fallback func(ParseError) T) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		result := p.Match(ctx)
		if result.Kind != Failed {
			return result
		}
		err := result.Err
		if err == nil {
			err = failureAt(ctx, max(ctx.Pos, result.End), "invalid syntax")
		}
		ctx.Diagnostics = append(ctx.Diagnostics, err)
		// A child can report a furthest error position after restoring its cursor.
		if ctx.Pos < result.End {
			ctx.Pos = min(result.End, uint32(len(ctx.Tokens)))
		}
		recovered := strategy.Recover(ctx)
		return Match[T]{Kind: Matched, Value: fallback(err), Start: result.Start, End: max(result.End, recovered.End)}
	})
}
func Recover[T any](p Pattern[T], strategy RecoveryStrategy) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		result := p.Match(ctx)
		if result.Kind == Failed {
			if result.Err == nil {
				result.Err = failureAt(ctx, max(ctx.Pos, result.End), "invalid syntax")
			}
			ctx.Diagnostics = append(ctx.Diagnostics, result.Err)
			if ctx.Pos < result.End {
				ctx.Pos = min(result.End, uint32(len(ctx.Tokens)))
			}
			strategy.Recover(ctx)
		}
		return result
	})
}
