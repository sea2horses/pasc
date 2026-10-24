package neoparser

import (
	"pseint-compiled/internal/ast"
	"pseint-compiled/internal/lexer"
	"pseint-compiled/internal/semantic"
)

func UnaryOperator() Pattern[*ast.Operator] {
	return Map(
		Locate(
			OneOf(
				If(
					Tok(lexer.MINUS),
					semantic.SUBTRACT,
				),
				If(
					Tok(lexer.EX_MARK),
					semantic.NOT,
				),
			),
		),
		func(o Located[semantic.OperatorType]) *ast.Operator {
			return &ast.Operator{
				NodeInfo: ast.NodeInfo{
					Span: o.Span,
				},
				Type: o.Value,
			}
		},
	)
}

func BinaryOperator() Pattern[*ast.Operator] {
	return Map(
		Locate(
			OneOf(
				If(
					Tok(lexer.PLUS),
					semantic.ADD,
				),
				If(
					Tok(lexer.MINUS),
					semantic.SUBTRACT,
				),
				If(
					Tok(lexer.ASTERISK),
					semantic.MULTIPLY,
				),
				If(
					Tok(lexer.SLASH),
					semantic.DIVIDE,
				),
				If(
					OneOf(
						AsAny(
							Tok(lexer.MODULO),
						),
						AsAny(
							Kw(lexer.MOD),
						),
					),
					semantic.MODULO,
				),
				If(
					Tok(lexer.R_ANGLE),
					semantic.GREATER,
				),
				If(
					Seq2(
						Tok(lexer.R_ANGLE),
						Tok(lexer.EQUALS),
					),
					semantic.GREATER_EQ,
				),
				If(
					Tok(lexer.L_ANGLE),
					semantic.LESSER,
				),
				If(
					Seq2(
						Tok(lexer.L_ANGLE),
						Tok(lexer.EQUALS),
					),
					semantic.LESSER_EQ,
				),
				If(
					Seq2(
						Tok(lexer.EQUALS),
						Tok(lexer.EQUALS),
					),
					semantic.EQUALS,
				),
				If(
					Seq2(
						Tok(lexer.EX_MARK),
						Tok(lexer.EQUALS),
					),
					semantic.NOT_EQUALS,
				),
				If(
					OneOf(
						AsAny(
							Seq2(
								Tok(lexer.AMPERSAND),
								Tok(lexer.AMPERSAND),
							),
						),
						AsAny(
							Tok(lexer.AMPERSAND),
						),
					),
					semantic.AND,
				),
				If(
					OneOf(
						AsAny(
							Seq2(
								Tok(lexer.PIPE),
								Tok(lexer.PIPE),
							),
						),
						AsAny(
							Tok(lexer.PIPE),
						),
					),
					semantic.OR,
				),
			),
		),
		func(o Located[semantic.OperatorType]) *ast.Operator {
			return &ast.Operator{
				NodeInfo: ast.NodeInfo{
					Span: o.Span,
				},
				Type: o.Value,
			}
		},
	)
}
