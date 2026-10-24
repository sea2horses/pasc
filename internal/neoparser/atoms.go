package neoparser

import (
	"pseint-compiled/internal/ast"
	"pseint-compiled/internal/lexer"
	"pseint-compiled/internal/models"
)

func Atom() Pattern[ast.Expr] {
	return OneOf(
		Identifier(),
		StringLiteral(),
		NumberLiteral(),
		BooleanLiteral(),
		ParenthesizedExpression(),
	)
}

func StringLiteral() Pattern[ast.Expr] {
	return Map(
		Tok(lexer.STRING_LITERAL),
		func(tok lexer.Token) ast.Expr {
			return &ast.StringLiteral{
				Content: tok.Value,
				NodeInfo: ast.NodeInfo{
					Span: tok.Span,
				},
			}
		},
	)
}

func RawNumber() Pattern[string] {
	return Map(
		Tok(lexer.NUMBER_LITERAL),
		func(tok lexer.Token) string {
			return tok.Value
		},
	)
}

func NumberLiteral() Pattern[ast.Expr] {
	return MapWithLocation(
		Seq2(
			RawNumber(),
			Optional(
				Right(
					Tok(lexer.DOT),
					RawNumber(),
				),
			),
		),
		func(p Pair[string, OptionalValue[string]], span models.Span) ast.Expr {
			return &ast.NumberLiteral{
				Int:  p.First,
				Frac: p.Second.Or(""),
				NodeInfo: ast.NodeInfo{
					Span: span,
				},
			}
		},
	)
}

func BooleanLiteral() Pattern[ast.Expr] {
	return Map(
		Tok(lexer.BOOLEAN_LITERAL),
		func(tok lexer.Token) ast.Expr {
			value, _ := lexer.MapToBool([]rune(tok.Value))
			return &ast.BoolLiteral{
				Value: value,
				NodeInfo: ast.NodeInfo{
					Span: tok.Span,
				},
			}
		},
	)
}

func Identifier() Pattern[ast.Expr] {
	return Map(Tok(lexer.IDENTIFIER), func(t lexer.Token) ast.Expr {
		return &ast.Variable{
			Name: t.Value,
			NodeInfo: ast.NodeInfo{
				Span: t.Span,
			},
		}
	})
}

func ParenthesizedExpression() Pattern[ast.Expr] {
	return Between(
		Tok(lexer.L_PARENTHESES),
		Expression(),
		Tok(lexer.R_PARENTHESES),
	)
}
