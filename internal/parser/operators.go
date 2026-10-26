package parser

import (
	"pascc/internal/ast"
	"pascc/internal/lexer"
	"pascc/internal/semantic"
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
				If(
					Kw(lexer.NO),
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
					Seq2(Tok(lexer.ASTERISK), Tok(lexer.ASTERISK)),
					semantic.POWER,
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
					OneOf(
						Seq2(
							/* also done as <> */
							Tok(lexer.L_ANGLE),
							Tok(lexer.R_ANGLE),
						),
						Seq2(
							Tok(lexer.EX_MARK),
							Tok(lexer.EQUALS),
						),
					),
					semantic.NOT_EQUALS,
				),
				If(
					Seq2(
						Tok(lexer.R_ANGLE),
						Tok(lexer.EQUALS),
					),
					semantic.GREATER_EQ,
				),
				If(
					Tok(lexer.R_ANGLE),
					semantic.GREATER,
				),
				If(
					Seq2(
						Tok(lexer.L_ANGLE),
						Tok(lexer.EQUALS),
					),
					semantic.LESSER_EQ,
				),
				If(
					Tok(lexer.L_ANGLE),
					semantic.LESSER,
				),
				If(
					OneOf(
						AsAny(Seq2(
							Tok(lexer.EQUALS),
							Tok(lexer.EQUALS),
						)),
						AsAny(Tok(lexer.EQUALS)),
					),
					semantic.EQUALS,
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
						AsAny(
							Kw(lexer.Y),
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
						AsAny(
							Kw(lexer.O),
						),
					),
					semantic.OR,
				),
				If(
					Tok(lexer.UPTICK),
					semantic.POWER,
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
