package parser

import (
	"pascc/internal/ast"
	"pascc/internal/lexer"
	"pascc/internal/models"
	"pascc/internal/semantic"
)

func Statement() Pattern[ast.Stmt] {
	return OneOf(
		Declaration(),
		Assignment(),
		ExpressionStatement(),
		Write(),
		Read(),
		IfStmt(),
		While(),
		For(),
		Switch(),
		Dimension(),
		Timeout(),
		ClearScreen(),
	)
}

// ExpressionStatement permits expressions such as standalone function calls.
// Assignment is tried first so an assignment's required RHS cannot be skipped.
func ExpressionStatement() Pattern[ast.Stmt] {
	return MapWithLocation(Expression(), func(expr ast.Expr, span models.Span) ast.Stmt {
		return &ast.ExprStmt{NodeInfo: ast.NodeInfo{Span: span}, Expr: expr}
	})
}

// StatementBlock consumes line breaks and leaves structural boundaries to its caller.
func StatementBlock() Pattern[[]ast.Stmt] {
	return statementBlock(false)
}

var blockBoundaries = []lexer.Keyword{
	lexer.FINALGORITMO, lexer.SINO, lexer.FINSI, lexer.FINMIENTRAS,
	lexer.FINPARA, lexer.FINSEGUN, lexer.DE,
}

func atBlockBoundary(ctx *Context, cases bool) bool {
	if ctx.EOF() {
		return true
	}
	for _, kw := range blockBoundaries {
		if Peek(Kw(kw)).Match(ctx).OK() {
			return true
		}
	}
	return cases && Peek(CaseHeader()).Match(ctx).OK()
}

func skipNewlines(ctx *Context) {
	for Newline().Match(ctx).OK() {
	}
}

func statementBlock(cases bool) Pattern[[]ast.Stmt] {
	return PatternFunc[[]ast.Stmt](func(ctx *Context) Match[[]ast.Stmt] {
		start := ctx.Pos
		// Headers and statements must end on their own line.
		if !ctx.EOF() && !Newline().Match(ctx).OK() {
			ctx.Diagnostics = append(ctx.Diagnostics, failureAt(ctx, ctx.Pos, ExpectedNewline.Expected()))
		}
		skipNewlines(ctx)
		stmts := []ast.Stmt{}
		recovery := Synchronize{Tokens: []lexer.TokenType{lexer.NEWLINE}, Keywords: blockBoundaries}
		for !atBlockBoundary(ctx, cases) {
			before := ctx.Pos
			result := Recover(Expected(Statement(), ExpectedStatement), recovery).Match(ctx)
			if result.OK() {
				stmts = append(stmts, result.Value)
			}
			if ctx.EOF() || atBlockBoundary(ctx, cases) {
				break
			}
			if result.OK() && !Newline().Match(ctx).OK() {
				ctx.Diagnostics = append(ctx.Diagnostics, failureAt(ctx, ctx.Pos, ExpectedNewline.Expected()))
				recovery.Recover(ctx)
			}
			skipNewlines(ctx)
			if ctx.Pos == before {
				ctx.Advance()
			}
		}
		return Match[[]ast.Stmt]{Kind: Matched, Value: stmts, Start: start, End: ctx.Pos}
	})
}

func MainFunction() Pattern[*ast.MainFunction] {
	return Map(
		Seq2(
			Locate(
				Right(
					Kw(lexer.ALGORITMO),
					Expected(Name(), ExpectedToken{Kind: lexer.IDENTIFIER}),
				),
			),
			Left(
				StatementBlock(),
				Expected(Kw(lexer.FINALGORITMO), ExpectedKeyword{Keyword: lexer.FINALGORITMO}),
			),
		),
		func(p Pair[Located[string], []ast.Stmt]) *ast.MainFunction {
			return &ast.MainFunction{
				NodeInfo: ast.NodeInfo{
					Span: p.First.Span,
				},
				Name:  p.First.Value,
				Stmts: p.Second,
			}
		},
	)
}

func Declaration() Pattern[ast.Stmt] {
	return Map(
		Locate(
			After(
				Kw(lexer.DEFINIR),
				Seq2(
					SepBy1(
						Expected(Name(), ExpectedToken{Kind: lexer.IDENTIFIER}),
						Tok(lexer.COMMA),
					),
					Right(
						Expected(Kw(lexer.COMO), ExpectedKeyword{Keyword: lexer.COMO}),
						Expected(Type(), ExpectedType),
					),
				),
			),
		),
		func(p Located[Pair[[]string, *ast.TypeRef]]) ast.Stmt {
			return &ast.Declaration{
				NodeInfo: ast.NodeInfo{
					Span: p.Span,
				},
				Names: p.Value.First,
				Type:  p.Value.Second,
			}
		},
	)
}

func Assignment() Pattern[ast.Stmt] {
	return MapWithLocation(
		Seq2(
			Expression(),
			Right(
				OneOf(
					AsAny(Tok(lexer.EQUALS)),
					AsAny(
						Seq2(
							/* -> */
							Tok(lexer.MINUS),
							Tok(lexer.R_ANGLE),
						),
					),
				),
				Expected(Expression(), ExpectedExpression),
			),
		),
		func(p Pair[ast.Expr, ast.Expr], span models.Span) ast.Stmt {
			return &ast.Assignment{
				NodeInfo: ast.NodeInfo{
					Span: span,
				},
				Target:  p.First,
				Content: p.Second,
			}
		},
	)
}

func Dimension() Pattern[ast.Stmt] {
	return MapWithLocation(
		After(
			Kw(lexer.DIMENSIONAR),
			Seq2(
				Expected(Name(), ExpectedToken{Kind: lexer.IDENTIFIER}),
				Between(
					Expected(Tok(lexer.L_BRACKET), ExpectedToken{Kind: lexer.L_BRACKET}),
					SepBy1(
						Expected(Expression(), ExpectedExpression),
						Tok(lexer.COMMA),
					),
					Expected(Tok(lexer.R_BRACKET), ExpectedToken{Kind: lexer.R_BRACKET}),
				),
			),
		),
		func(p Pair[string, []ast.Expr], span models.Span) ast.Stmt {
			return &ast.Dimension{
				NodeInfo: ast.NodeInfo{
					Span: span,
				},
				Name:       p.First,
				Dimensions: p.Second,
			}
		},
	)
}

func Timeout() Pattern[ast.Stmt] {
	return MapWithLocation(
		After(
			Kw(lexer.ESPERAR),
			Seq2(
				Expected(Expression(), ExpectedExpression),
				Expected(TimeUnit(), LiteralExpectations("expected time unit")),
			),
		),
		func(p Pair[ast.Expr, semantic.TimeUnit], span models.Span) ast.Stmt {
			return &ast.TimeOut{
				NodeInfo: ast.NodeInfo{
					Span: span,
				},
				Amount: p.First,
				Unit:   p.Second,
			}
		},
	)
}

func ClearScreen() Pattern[ast.Stmt] {
	return Map(
		Locate(
			AsAny(Seq2(
				Kw(lexer.BORRAR),
				Expected(Kw(lexer.PANTALLA), ExpectedKeyword{Keyword: lexer.PANTALLA}),
			)),
		),
		func(p Located[any]) ast.Stmt {
			return &ast.ClearScreen{
				NodeInfo: ast.NodeInfo{
					Span: p.Span,
				},
			}
		},
	)
}
