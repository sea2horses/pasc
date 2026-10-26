package parser

import (
	"pascc/internal/ast"
	"pascc/internal/lexer"
	"pascc/internal/models"
	"pascc/internal/semantic"
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
					Expected(EndKeyword(lexer.FINSI, lexer.SI), ExpectedKeyword{Keyword: lexer.FINSI}),
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
	return PatternFunc[*ast.Else](func(ctx *Context) Match[*ast.Else] {
		return elseStmtPattern().Match(ctx)
	})
}

func elseStmtPattern() Pattern[*ast.Else] {
	elseIf := Map(
		After(
			Kw(lexer.SI),
			Left(
				Seq3(
					Locate(Left(Expected(Expression(), ExpectedExpression), Optional(Kw(lexer.ENTONCES)))),
					StatementBlock(),
					Optional(ElseStmt()),
				),
				Expected(EndKeyword(lexer.FINSI, lexer.SI), ExpectedKeyword{Keyword: lexer.FINSI}),
			),
		),
		func(branch Trio[Located[ast.Expr], []ast.Stmt, OptionalValue[*ast.Else]]) []ast.Stmt {
			return []ast.Stmt{&ast.If{
				NodeInfo:  ast.NodeInfo{Span: branch.First.Span},
				Condition: branch.First.Value,
				Stmts:     branch.Second,
				Else:      branch.Third.Or(nil),
			}}
		},
	)
	body := OneOf(elseIf, StatementBlock())
	return Map(
		Seq2(Locate(Kw(lexer.SINO)), Left(body, Many(Newline()))),
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
				Expected(EndKeyword(lexer.FINMIENTRAS, lexer.MIENTRAS), ExpectedKeyword{Keyword: lexer.FINMIENTRAS}),
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
			Locate(Left(ForHeader(), Optional(Kw(lexer.HACER)))),
			Left(
				StatementBlock(),
				Expected(EndKeyword(lexer.FINPARA, lexer.PARA), ExpectedKeyword{Keyword: lexer.FINPARA}),
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

func ForHeader() Pattern[Trio[ast.Stmt, ast.Expr, OptionalValue[ast.Expr]]] {
	start := Right(Kw(lexer.PARA), Expected(ForStart(), ExpectedStatement))
	standard := Seq3(
		start,
		Right(Kw(lexer.HASTA), Expected(Expression(), ExpectedExpression)),
		Optional(Right(
			Seq2(Kw(lexer.CON), Expected(Kw(lexer.PASO), ExpectedKeyword{Keyword: lexer.PASO})),
			Expected(Expression(), ExpectedExpression),
		)),
	)
	stepBeforeEnd := Map(
		Seq3(
			start,
			Right(
				Seq2(Kw(lexer.CON), Expected(Kw(lexer.PASO), ExpectedKeyword{Keyword: lexer.PASO})),
				Expected(Expression(), ExpectedExpression),
			),
			Right(Kw(lexer.HASTA), Expected(Expression(), ExpectedExpression)),
		),
		func(parts Trio[ast.Stmt, ast.Expr, ast.Expr]) Trio[ast.Stmt, ast.Expr, OptionalValue[ast.Expr]] {
			return Trio[ast.Stmt, ast.Expr, OptionalValue[ast.Expr]]{
				First:  parts.First,
				Second: parts.Third,
				Third:  OptionalValue[ast.Expr]{Value: parts.Second, Some: true},
			}
		},
	)
	return OneOf(standard, stepBeforeEnd)
}

func ForStart() Pattern[ast.Stmt] {
	return OneOf(
		Assignment(),
		MapWithLocation(Name(), func(name string, span models.Span) ast.Stmt {
			variable := &ast.Variable{NodeInfo: ast.NodeInfo{Span: span}, Name: name}
			return &ast.Assignment{NodeInfo: ast.NodeInfo{Span: span}, Target: variable, Content: variable}
		}),
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
			Expected(EndKeyword(lexer.FINSEGUN, lexer.SEGUN), ExpectedKeyword{Keyword: lexer.FINSEGUN}),
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

func CaseHeader() Pattern[[]ast.Expr] {
	return Map(
		Right(
			Optional(Kw(lexer.CASO)),
			Left(SepBy1(Expression(), Tok(lexer.COMMA)), Tok(lexer.COLON)),
		),
		func(values []ast.Expr) []ast.Expr {
			var flattened []ast.Expr
			for _, value := range values {
				flattenCaseOrLabels(value, &flattened)
			}
			return flattened
		},
	)
}

func flattenCaseOrLabels(expr ast.Expr, into *[]ast.Expr) {
	if operation, ok := expr.(*ast.BinaryOperation); ok && operation.Op.Type == semantic.OR {
		flattenCaseOrLabels(operation.LHS, into)
		flattenCaseOrLabels(operation.RHS, into)
		return
	}
	*into = append(*into, expr)
}

func DefaultHeader() Pattern[any] {
	return AsAny(Left(Seq3(
		Kw(lexer.DE),
		Expected(Kw(lexer.OTRO), ExpectedKeyword{Keyword: lexer.OTRO}),
		Expected(Kw(lexer.MODO), ExpectedKeyword{Keyword: lexer.MODO}),
	), Optional(Tok(lexer.COLON))))
}

func EndKeyword(compound, tail lexer.Keyword) Pattern[lexer.Keyword] {
	return OneOf(
		Kw(compound),
		Map(
			Seq2(Kw(lexer.FIN), Kw(tail)),
			func(Pair[lexer.Keyword, lexer.Keyword]) lexer.Keyword { return compound },
		),
	)
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
		func(c Pair[Located[[]ast.Expr], []ast.Stmt]) *ast.Case {
			return &ast.Case{
				NodeInfo: ast.NodeInfo{
					Span: c.First.Span,
				},
				Clauses: c.First.Value,
				Stmts:   c.Second,
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
		for !atBlockBoundary(ctx, false, false) {
			before := ctx.Pos
			result := Recover(Expected(SwitchCase(), LiteralExpectations("expected switch case")), recovery).Match(ctx)
			if result.OK() {
				cases = append(cases, result.Value)
			}
			skipNewlines(ctx)
			if ctx.Pos == before && !atBlockBoundary(ctx, false, false) {
				ctx.Advance()
			}
		}
		return Match[[]*ast.Case]{Kind: Matched, Value: cases, Start: start, End: ctx.Pos}
	})
}

func DoWhile() Pattern[ast.Stmt] {
	until := Map(
		After(
			Kw(lexer.REPETIR),
			Seq2(
				statementBlockWithPostTestWhile(false, true),
				Locate(Map(
					Seq2(
						OneOf(
							Map(Seq2(Kw(lexer.HASTA), Kw(lexer.QUE)), func(Pair[lexer.Keyword, lexer.Keyword]) bool { return true }),
							Map(Seq2(Kw(lexer.MIENTRAS), Kw(lexer.QUE)), func(Pair[lexer.Keyword, lexer.Keyword]) bool { return false }),
						),
						Expression(),
					),
					func(p Pair[bool, ast.Expr]) Pair[bool, ast.Expr] { return p },
				)),
			),
		),
		func(p Pair[[]ast.Stmt, Located[Pair[bool, ast.Expr]]]) ast.Stmt {
			return &ast.DoWhile{
				NodeInfo:  ast.NodeInfo{Span: p.Second.Span},
				Stmts:     p.First,
				Condition: p.Second.Value.Second,
				Until:     p.Second.Value.First,
			}
		},
	)
	while := Map(
		After(
			Kw(lexer.HACER),
			Seq2(
				statementBlockWithPostTestWhile(false, true),
				Locate(Right(
					Seq2(
						Expected(Kw(lexer.MIENTRAS), ExpectedKeyword{Keyword: lexer.MIENTRAS}),
						Expected(Kw(lexer.QUE), ExpectedKeyword{Keyword: lexer.QUE}),
					),
					Left(Expression(), Optional(Kw(lexer.HACER))),
				)),
			),
		),
		func(p Pair[[]ast.Stmt, Located[ast.Expr]]) ast.Stmt {
			return &ast.DoWhile{
				NodeInfo:  ast.NodeInfo{Span: p.Second.Span},
				Stmts:     p.First,
				Condition: p.Second.Value,
				Until:     false,
			}
		},
	)
	return OneOf(until, while)
}
