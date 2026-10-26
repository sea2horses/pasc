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
		DoWhile(),
		For(),
		Switch(),
		Dimension(),
		WaitKey(),
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
	lexer.FINALGORITMO, lexer.FIN, lexer.SINO, lexer.FINSI, lexer.FINMIENTRAS,
	lexer.FINPARA, lexer.FINSEGUN, lexer.DE,
	lexer.HASTA, lexer.FINPROCESO,
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

func Skippable() Pattern[any] {
	return OneOf(
		AsAny(Newline()),
		AsAny(Tok(lexer.SEMICOLON)),
	)
}

func skippables(ctx *Context) {
	for Skippable().Match(ctx).OK() {
	}
}

func skipSemicolons(ctx *Context) {
	for Tok(lexer.SEMICOLON).Match(ctx).OK() {
	}
}

func skipNewlines(ctx *Context) {
	for Newline().Match(ctx).OK() {
	}
}

func BlockHeaderNewline() Pattern[lexer.Token] {
	return PatternFunc[lexer.Token](func(ctx *Context) Match[lexer.Token] {
		skipSemicolons(ctx)
		return Expected(Newline(), ExpectedNewline).Match(ctx)
	})
}

func statementBlock(cases bool) Pattern[[]ast.Stmt] {
	return PatternFunc[[]ast.Stmt](func(ctx *Context) Match[[]ast.Stmt] {
		start := ctx.Pos
		// Headers and statements must end on their own line.
		skipSemicolons(ctx)
		if !ctx.EOF() && !Newline().Match(ctx).OK() {
			ctx.Diagnostics = append(ctx.Diagnostics, failureAt(ctx, ctx.Pos, ExpectedNewline.Expected()))
		}
		skippables(ctx)
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
			// A semicolon separates statements, including on the same line. Without
			// one, statements must remain newline-separated.
			if Tok(lexer.SEMICOLON).Match(ctx).OK() {
				skipSemicolons(ctx)
				if Newline().Match(ctx).OK() {
					skipNewlines(ctx)
				}
			} else if Newline().Match(ctx).OK() {
				skipNewlines(ctx)
			} else if result.OK() {
				ctx.Diagnostics = append(ctx.Diagnostics, failureAt(ctx, ctx.Pos, ExpectedNewline.Expected()))
				recovery.Recover(ctx)
			}
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
					OneOf(
						Kw(lexer.ALGORITMO),
						Kw(lexer.PROCESO),
					),
					Expected(Name(), ExpectedToken{Kind: lexer.IDENTIFIER}),
				),
			),
			Left(
				StatementBlock(),
				Expected(
					OneOf(
						EndKeyword(lexer.FINALGORITMO, lexer.ALGORITMO),
						EndKeyword(lexer.FINPROCESO, lexer.PROCESO),
					),
					ExpectedKeyword{Keyword: lexer.FINALGORITMO},
				),
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
			Operand(),
			Right(
				OneOf(
					AsAny(Tok(lexer.EQUALS)),
					AsAny(
						Seq2(
							/* <- */
							Tok(lexer.L_ANGLE),
							Tok(lexer.MINUS),
						),
					),
					AsAny(
						Seq2(
							/* := */
							Tok(lexer.COLON),
							Tok(lexer.EQUALS),
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
			OneOf(
				Kw(lexer.DIMENSIONAR),
				Kw(lexer.DIMENSION),
			),
			SepBy1(DimensionDeclaration(), Tok(lexer.COMMA)),
		),
		func(declarations []*ast.Dimension, span models.Span) ast.Stmt {
			return &ast.DimensionGroup{
				NodeInfo: ast.NodeInfo{
					Span: span,
				},
				Declarations: declarations,
			}
		},
	)
}

func DimensionDeclaration() Pattern[*ast.Dimension] {
	return MapWithLocation(
		Seq2(
			Expected(Name(), ExpectedToken{Kind: lexer.IDENTIFIER}),
			Many1(OneOf(
				Between(
					Tok(lexer.L_BRACKET),
					SepBy1(Expected(Expression(), ExpectedExpression), Tok(lexer.COMMA)),
					Tok(lexer.R_BRACKET),
				),
				Between(
					Tok(lexer.L_PARENTHESES),
					SepBy1(Expected(Expression(), ExpectedExpression), Tok(lexer.COMMA)),
					Tok(lexer.R_PARENTHESES),
				),
			)),
		),
		func(p Pair[string, [][]ast.Expr], span models.Span) *ast.Dimension {
			dimensions := make([]ast.Expr, 0)
			for _, group := range p.Second {
				dimensions = append(dimensions, group...)
			}
			return &ast.Dimension{
				NodeInfo:   ast.NodeInfo{Span: span},
				Name:       p.First,
				Dimensions: dimensions,
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

func WaitKey() Pattern[ast.Stmt] {
	return MapWithLocation(
		Seq2(Kw(lexer.ESPERAR), Kw(lexer.TECLA)),
		func(_ Pair[lexer.Keyword, lexer.Keyword], span models.Span) ast.Stmt {
			return &ast.WaitKey{NodeInfo: ast.NodeInfo{Span: span}}
		},
	)
}

func ClearScreen() Pattern[ast.Stmt] {
	return Map(
		Locate(
			AsAny(After(
				OneOf(
					Kw(lexer.BORRAR),
					Kw(lexer.LIMPIAR),
				),
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
