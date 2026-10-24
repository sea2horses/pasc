package neoparser

import (
	"pseint-compiled/internal/ast"
	"pseint-compiled/internal/lexer"
)

func IfStmt() Pattern[ast.Stmt] {
	return Map(
		After(
			Kw(lexer.SI),
			Seq3(
				Locate(
					Left(
						Expression(),
						Optional(
							Kw(lexer.ENTONCES),
						),
					),
				),
				StatementBlock(),
				Left(
					Optional(
						ElseStmt(),
					),
					Seq2(
						Newline(),
						Kw(lexer.FINSI),
					),
				),
			),
		),
		func(p Trio[Located[ast.Expr], []ast.Stmt, OptionalValue[*ast.Else]]) ast.Stmt {
			return &ast.If{
				NodeInfo: ast.NodeInfo{
					Span: p.First.Span,
				},
				Condition: p.First.Value,
				Stmts:     p.Second,
				Else:      p.Third.Or(nil),
			}
		},
	)
}

func ElseStmt() Pattern[*ast.Else] {
	return Map(
		Seq2(
			Locate(
				Kw(lexer.SINO),
			),
			StatementBlock(),
		),
		func(p Pair[Located[lexer.Keyword], []ast.Stmt]) *ast.Else {
			return &ast.Else{
				NodeInfo: ast.NodeInfo{
					Span: p.First.Span,
				},
				Stmts: p.Second,
			}
		},
	)
}

func While() Pattern[ast.Stmt] {
	return Map(
		Seq2(
			Locate(
				Between(
					Seq2(
						Kw(lexer.MIENTRAS),
						Optional(Kw(lexer.QUE)),
					),
					Expression(),
					Optional(
						Kw(lexer.HACER),
					),
				),
			),
			Left(
				StatementBlock(),
				Seq2(
					Newline(),
					Kw(lexer.FINMIENTRAS),
				),
			),
		),
		func(p Pair[Located[ast.Expr], []ast.Stmt]) ast.Stmt {
			return &ast.While{
				NodeInfo: ast.NodeInfo{
					Span: p.First.Span,
				},
				Condition: p.First.Value,
				Stmts:     p.Second,
			}
		},
	)
}

func For() Pattern[ast.Stmt] {
	return Map(
		Seq2(
			Locate(
				Left(
					Seq3(
						Right(
							Kw(lexer.PARA),
							Assignment(),
						),
						Right(
							Kw(lexer.HASTA),
							Expression(),
						),
						Optional(
							Right(
								Seq2(
									Kw(lexer.CON),
									Kw(lexer.PASO),
								),
								Expression(),
							),
						),
					),
					Optional(Kw(lexer.HACER)),
				),
			),
			Left(
				StatementBlock(),
				Seq2(
					Newline(),
					Kw(lexer.FINPARA),
				),
			),
		),
		func(t Pair[Located[Trio[ast.Stmt, ast.Expr, OptionalValue[ast.Expr]]], []ast.Stmt]) ast.Stmt {
			header := t.First.Value

			return &ast.For{
				NodeInfo: ast.NodeInfo{
					Span: t.First.Span,
				},
				Start: header.First.(*ast.Assignment),
				Until: header.Second,
				Step:  header.Third.Or(nil),
				Stmts: t.Second,
			}
		},
	)
}

func Switch() Pattern[ast.Stmt] {
	return Map(
		Left(
			Seq3(
				Locate(
					Between(
						Kw(lexer.SEGUN),
						Expression(),
						Optional(Kw(lexer.HACER)),
					),
				),
				Many(
					SwitchCase(),
				),
				Optional(
					Right(
						DefaultHeader(),
						StatementBlock(),
					),
				),
			),
			Kw(lexer.FINSEGUN),
		),
		func(t Trio[Located[ast.Expr], []*ast.Case, OptionalValue[[]ast.Stmt]]) ast.Stmt {
			return &ast.Switch{
				NodeInfo: ast.NodeInfo{
					Span: t.First.Span,
				},
				Base:    t.First.Value,
				Cases:   t.Second,
				Default: t.Third.Or(nil),
			}
		},
	)
}

func CaseHeader() Pattern[ast.Expr] {
	return Left(
		Expression(),
		Tok(lexer.COLON),
	)
}

func DefaultHeader() Pattern[any] {
	return AsAny(Seq2(
		Seq3(
			Kw(lexer.DE),
			Kw(lexer.OTRO),
			Kw(lexer.MODO),
		),
		Tok(lexer.COLON),
	))
}

func SwitchCase() Pattern[*ast.Case] {
	return Map(
		Seq2(
			/* Header */
			Locate(
				CaseHeader(),
			),
			Until(
				Left(
					Statement(),
					Newline(),
				),
				OneOf(
					AsAny(CaseHeader()),
					DefaultHeader(),
				),
			),
		),
		func(c Pair[Located[ast.Expr], []ast.Stmt]) *ast.Case {
			return &ast.Case{
				NodeInfo: ast.NodeInfo{
					Span: c.First.Span,
				},
				Clause: c.First.Value,
				Stmts:  c.Second,
			}
		},
	)
}
