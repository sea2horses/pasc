package parser

import (
	"pascc/internal/ast"
	"pascc/internal/lexer"
)

func IfStmt() Pattern[ast.Stmt] {
	return Map(
		After(
			Kw(lexer.SI),
			Seq3(
				Locate(
					Left(
						Expected(Expression(), ExpectedExpression),
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
					Expected(Kw(lexer.FINSI), ExpectedKeyword{Keyword: lexer.FINSI}),
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
					Expected(Expression(), ExpectedExpression),
					Optional(
						Kw(lexer.HACER),
					),
				),
			),
			Left(
				StatementBlock(),
				Expected(Kw(lexer.FINMIENTRAS), ExpectedKeyword{Keyword: lexer.FINMIENTRAS}),
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
							Expected(Assignment(), ExpectedStatement),
						),
						Right(
							Expected(Kw(lexer.HASTA), ExpectedKeyword{Keyword: lexer.HASTA}),
							Expected(Expression(), ExpectedExpression),
						),
						Optional(
							Right(
								Seq2(
									Kw(lexer.CON),
									Expected(Kw(lexer.PASO), ExpectedKeyword{Keyword: lexer.PASO}),
								),
								Expected(Expression(), ExpectedExpression),
							),
						),
					),
					Optional(Kw(lexer.HACER)),
				),
			),
			Left(
				StatementBlock(),
				Expected(Kw(lexer.FINPARA), ExpectedKeyword{Keyword: lexer.FINPARA}),
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
						Expected(Expression(), ExpectedExpression),
						Optional(Kw(lexer.HACER)),
					),
				),
				Right(
					BlockHeaderNewline(),
					Right(Many(Newline()), switchCases()),
				),
				Optional(
					Right(
						DefaultHeader(),
						StatementBlock(),
					),
				),
			),
			Expected(Kw(lexer.FINSEGUN), ExpectedKeyword{Keyword: lexer.FINSEGUN}),
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
			Expected(Kw(lexer.OTRO), ExpectedKeyword{Keyword: lexer.OTRO}),
			Expected(Kw(lexer.MODO), ExpectedKeyword{Keyword: lexer.MODO}),
		),
		Expected(Tok(lexer.COLON), ExpectedToken{Kind: lexer.COLON}),
	))
}

func SwitchCase() Pattern[*ast.Case] {
	return Map(
		Seq2(
			/* Header */
			Locate(
				CaseHeader(),
			),
			statementBlock(true),
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

// A malformed case header must not prevent later cases from being parsed.
func switchCases() Pattern[[]*ast.Case] {
	return PatternFunc[[]*ast.Case](func(ctx *Context) Match[[]*ast.Case] {
		start := ctx.Pos
		cases := []*ast.Case{}
		recovery := Synchronize{Tokens: []lexer.TokenType{lexer.NEWLINE}, Keywords: blockBoundaries}
		for !atBlockBoundary(ctx, false) {
			before := ctx.Pos
			result := Recover(Expected(SwitchCase(), LiteralExpectations("expected switch case")), recovery).Match(ctx)
			if result.OK() {
				cases = append(cases, result.Value)
			}
			skipNewlines(ctx)
			if ctx.Pos == before && !atBlockBoundary(ctx, false) {
				ctx.Advance()
			}
		}
		return Match[[]*ast.Case]{Kind: Matched, Value: cases, Start: start, End: ctx.Pos}
	})
}

func DoWhile() Pattern[ast.Stmt] {
	return Map(
		After(
			Kw(lexer.REPETIR),
			Seq2(
				StatementBlock(),
				Locate(
					Right(
						Seq2(
							Expected(Kw(lexer.HASTA), ExpectedKeyword{Keyword: lexer.HASTA}),
							Expected(Kw(lexer.QUE), ExpectedKeyword{Keyword: lexer.QUE}),
						),
						Expression(),
					),
				),
			),
		),
		func(p Pair[[]ast.Stmt, Located[ast.Expr]]) ast.Stmt {
			return &ast.DoWhile{
				NodeInfo: ast.NodeInfo{
					Span: p.Second.Span,
				},
				Stmts:     p.First,
				Condition: p.Second.Value,
			}
		},
	)
}
