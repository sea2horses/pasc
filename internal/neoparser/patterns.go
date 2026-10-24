package neoparser

import (
	"pseint-compiled/internal/ast"
	"pseint-compiled/internal/lexer"
	"pseint-compiled/internal/semantic"
)

type Pattern[T any] interface {
	Match(*Context) Match[T]
}

type PatternFunc[T any] func(*Context) Match[T]

func (f PatternFunc[T]) Match(c *Context) Match[T] {
	return f(c)
}

/* Parse a token of the given type */
func Tok(expected lexer.TokenType) Pattern[lexer.Token] {
	return PatternFunc[lexer.Token](func(ctx *Context) Match[lexer.Token] {
		start := ctx.Mark()
		tok := ctx.Peek()

		if tok == nil {
			return Match[lexer.Token]{
				Kind:  NoMatch,
				Start: start,
				End:   start,
			}
		}

		if tok.Type != expected {
			return Match[lexer.Token]{
				Kind:  NoMatch,
				Start: start,
				End:   ctx.Pos,
			}
		}

		ctx.Advance()

		return Match[lexer.Token]{
			Kind:  Matched,
			Value: *tok,
			Start: start,
			End:   ctx.Pos,
		}
	})
}

func Kw(expected lexer.Keyword) Pattern[lexer.Keyword] {
	return PatternFunc[lexer.Keyword](func(ctx *Context) Match[lexer.Keyword] {
		start := ctx.Mark()
		tok := ctx.Peek()

		if tok == nil {
			return Match[lexer.Keyword]{
				Kind:  NoMatch,
				Start: start,
				End:   start,
			}
		}

		if tok.Type != lexer.KEYWORD {
			return Match[lexer.Keyword]{
				Kind:  NoMatch,
				Start: start,
				End:   ctx.Pos,
			}
		}

		kw, ok := lexer.MapToKeyword([]rune(tok.Value))

		if !ok || kw != expected {
			return Match[lexer.Keyword]{
				Kind:  NoMatch,
				Start: start,
				End:   ctx.Pos,
			}
		}

		ctx.Advance()

		return Match[lexer.Keyword]{
			Kind:  Matched,
			Value: kw,
			Start: start,
			End:   ctx.Pos,
		}
	})
}

func Name() Pattern[string] {
	return Map(Tok(lexer.IDENTIFIER), func(t lexer.Token) string {
		return t.Value
	})
}

func Newline() Pattern[lexer.Token] {
	return Tok(lexer.NEWLINE)
}

func Type() Pattern[*ast.TypeRef] {
	return OneOf(
		Map(
			Locate(
				OneOf(
					Kw(lexer.ENTERO),
					Kw(lexer.CADENA),
					Kw(lexer.REAL),
					Kw(lexer.LOGICO),
				),
			),
			func(kw Located[lexer.Keyword]) *ast.TypeRef {
				return &ast.TypeRef{
					NodeInfo: ast.NodeInfo{
						Span: kw.Span,
					},
					Name: kw.Value.String(),
				}
			},
		),
		Map(
			Locate(
				Name(),
			),
			func(name Located[string]) *ast.TypeRef {
				return &ast.TypeRef{
					NodeInfo: ast.NodeInfo{
						Span: name.Span,
					},
					Name: name.Value,
				}
			},
		),
	)
}

func TimeUnit() Pattern[semantic.TimeUnit] {
	return OneOf(
		If(
			Kw(lexer.MILISEGUNDOS),
			semantic.MILISEGUNDOS,
		),
		If(
			Kw(lexer.SEGUNDOS),
			semantic.SEGUNDOS,
		),
	)
}

/* Just in case */
func Custom[T any](
	fn func(*Context) Match[T],
) Pattern[T] {
	return PatternFunc[T](fn)
}
