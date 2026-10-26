package parser

import (
	"pascc/internal/ast"
	"pascc/internal/lexer"
	"pascc/internal/models"
)

func Write() Pattern[ast.Stmt] {
	return MapWithLocation(
		After(
			OneOf(
				Kw(lexer.ESCRIBIR),
				Kw(lexer.IMPRIMIR),
			),
			Seq3(
				Optional(Tok(lexer.COMMA)),
				Optional(writeNoNewline()),
				Seq2(ExpressionList(), Optional(OneOf(
					writeNoNewline(),
					Right(Tok(lexer.SEMICOLON), writeNoNewline()),
				))),
			),
		),
		func(p Trio[OptionalValue[lexer.Token], OptionalValue[any], Pair[[]ast.Expr, OptionalValue[any]]], span models.Span) ast.Stmt {
			return &ast.Write{
				NodeInfo: ast.NodeInfo{
					Span: span,
				},
				Print:   p.Third.First,
				Newline: !p.Second.Some && !p.Third.Second.Some,
			}
		},
	)
}

func writeNoNewline() Pattern[any] {
	return AsAny(Seq2(
		Kw(lexer.SIN),
		Expected(Kw(lexer.SALTAR), ExpectedKeyword{Keyword: lexer.SALTAR}),
	))
}

func Read() Pattern[ast.Stmt] {
	return MapWithLocation(
		After(
			Kw(lexer.LEER),
			ExpressionList(),
		),
		func(read []ast.Expr, span models.Span) ast.Stmt {
			return &ast.Read{
				NodeInfo: ast.NodeInfo{
					Span: span,
				},
				Into: read,
			}
		},
	)
}

func ExpressionList() Pattern[[]ast.Expr] {
	first := Left(Expected(Expression(), ExpectedExpression), Optional(Tok(lexer.COMMA)))
	more := Many(Left(Expression(), Optional(Tok(lexer.COMMA))))
	return Map(Seq2(first, more), func(items Pair[ast.Expr, []ast.Expr]) []ast.Expr {
		return append([]ast.Expr{items.First}, items.Second...)
	})
}
