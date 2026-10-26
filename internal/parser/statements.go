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
	lexer.FINPARA, lexer.FINSEGUN, lexer.FINFUNCION, lexer.DE,
	lexer.HASTA, lexer.FINPROCESO,
}

func atBlockBoundary(ctx *Context, cases, postTestWhile bool) bool {
	if ctx.EOF() {
		return true
	}
	for _, kw := range blockBoundaries {
		if Peek(Kw(kw)).Match(ctx).OK() {
			return true
		}
	}
	if Peek(EndKeyword(lexer.FINFUNCION, lexer.FUNCION)).Match(ctx).OK() {
		return true
	}
	if postTestWhile && atPostTestWhileTerminator(ctx) {
		return true
	}
	return cases && Peek(CaseHeader()).Match(ctx).OK()
}

func atPostTestWhileTerminator(ctx *Context) bool {
	if !Peek(Seq2(Kw(lexer.MIENTRAS), Kw(lexer.QUE))).Match(ctx).OK() {
		return false
	}
	first := ctx.Peek()
	for offset := uint32(0); ; offset++ {
		token := ctx.PeekAt(offset)
		if token == nil || token.Type == lexer.NEWLINE || token.Span.Start.Line != first.Span.Start.Line {
			break
		}
		if token.Type == lexer.KEYWORD {
			if kw, ok := lexer.MapToKeyword([]rune(token.Value)); ok && kw == lexer.HACER {
				return false
			}
		}
	}
	return true
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
	return statementBlockWithPostTestWhile(cases, false)
}

func statementBlockWithPostTestWhile(cases, postTestWhile bool) Pattern[[]ast.Stmt] {
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
		for !atBlockBoundary(ctx, cases, postTestWhile) {
			before := ctx.Pos
			result := Recover(Expected(Statement(), ExpectedStatement), recovery).Match(ctx)
			if result.OK() {
				stmts = append(stmts, result.Value)
			}
			if ctx.EOF() || atBlockBoundary(ctx, cases, postTestWhile) {
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

type functionHeader struct {
	returnVariable string
	name           string
	arguments      []string
	byReference    []bool
	hasReturn      bool
}

type functionArgument struct {
	name        string
	byReference bool
}

func FunctionArgument() Pattern[functionArgument] {
	return Map(
		Seq2(
			Expected(Name(), ExpectedToken{Kind: lexer.IDENTIFIER}),
			Optional(Right(
				Kw(lexer.POR),
				Expected(Kw(lexer.REFERENCIA), ExpectedKeyword{Keyword: lexer.REFERENCIA}),
			)),
		),
		func(p Pair[string, OptionalValue[lexer.Keyword]]) functionArgument {
			return functionArgument{name: p.First, byReference: p.Second.Some}
		},
	)
}

func FunctionArguments() Pattern[[]functionArgument] {
	return Between(
		Tok(lexer.L_PARENTHESES),
		SepBy(FunctionArgument(), Tok(lexer.COMMA)),
		Expected(Tok(lexer.R_PARENTHESES), ExpectedToken{Kind: lexer.R_PARENTHESES}),
	)
}

func FunctionHeader() Pattern[functionHeader] {
	return PatternFunc[functionHeader](func(ctx *Context) Match[functionHeader] {
		start := ctx.Checkpoint()
		keyword := Kw(lexer.FUNCION).Match(ctx)
		if keyword.Kind != Matched {
			return Match[functionHeader]{Kind: keyword.Kind, Start: start.Pos, End: keyword.End, Err: keyword.Err}
		}
		firstName := Expected(Name(), ExpectedToken{Kind: lexer.IDENTIFIER}).Match(ctx)
		if firstName.Kind != Matched {
			return Match[functionHeader]{Kind: firstName.Kind, Start: start.Pos, End: firstName.End, Err: firstName.Err}
		}

		header := functionHeader{}
		if Tok(lexer.L_ANGLE).Match(ctx).OK() {
			minus := Expected(Tok(lexer.MINUS), ExpectedToken{Kind: lexer.MINUS}).Match(ctx)
			if minus.Kind != Matched {
				return Match[functionHeader]{Kind: minus.Kind, Start: start.Pos, End: minus.End, Err: minus.Err}
			}
			name := Expected(Name(), ExpectedToken{Kind: lexer.IDENTIFIER}).Match(ctx)
			if name.Kind != Matched {
				return Match[functionHeader]{Kind: name.Kind, Start: start.Pos, End: name.End, Err: name.Err}
			}
			header.returnVariable = firstName.Value
			header.name = name.Value
			header.hasReturn = true
		} else {
			header.name = firstName.Value
		}

		arguments := FunctionArguments().Match(ctx)
		if arguments.Kind != Matched {
			return Match[functionHeader]{Kind: arguments.Kind, Start: start.Pos, End: arguments.End, Err: arguments.Err}
		}
		for _, argument := range arguments.Value {
			header.arguments = append(header.arguments, argument.name)
			header.byReference = append(header.byReference, argument.byReference)
		}
		return Match[functionHeader]{Kind: Matched, Value: header, Start: start.Pos, End: ctx.Pos}
	})
}

func FunctionDeclaration() Pattern[*ast.Function] {

	return Map(
		Locate(Seq3(
			FunctionHeader(),
			StatementBlock(),
			Expected(EndKeyword(lexer.FINFUNCION, lexer.FUNCION), ExpectedKeyword{Keyword: lexer.FINFUNCION}),
		)),
		func(p Located[Trio[functionHeader, []ast.Stmt, lexer.Keyword]]) *ast.Function {
			return &ast.Function{
				NodeInfo:       ast.NodeInfo{Span: p.Span},
				Name:           p.Value.First.name,
				ReturnVariable: p.Value.First.returnVariable,
				Arguments:      p.Value.First.arguments,
				ByReference:    p.Value.First.byReference,
				HasReturn:      p.Value.First.hasReturn,
				Stmts:          p.Value.Second,
			}
		},
	)
}

func Program() Pattern[*ast.MainFunction] {
	return PatternFunc[*ast.MainFunction](func(ctx *Context) Match[*ast.MainFunction] {
		start := ctx.Checkpoint()
		functions := []*ast.Function{}
		skipNewlines(ctx)
		for {
			before := ctx.Checkpoint()
			function := FunctionDeclaration().Match(ctx)
			if function.Kind == Failed {
				return Match[*ast.MainFunction]{Kind: Failed, Start: start.Pos, End: function.End, Err: function.Err}
			}
			if function.Kind == NoMatch {
				ctx.Restore(before)
				break
			}
			functions = append(functions, function.Value)
			skipNewlines(ctx)
		}

		main := MainFunction().Match(ctx)
		if main.Kind != Matched {
			return Match[*ast.MainFunction]{Kind: main.Kind, Start: start.Pos, End: main.End, Err: main.Err}
		}
		main.Value.Functions = functions
		skipNewlines(ctx)
		for {
			before := ctx.Checkpoint()
			function := FunctionDeclaration().Match(ctx)
			if function.Kind == Failed {
				return Match[*ast.MainFunction]{Kind: Failed, Start: start.Pos, End: function.End, Err: function.Err}
			}
			if function.Kind == NoMatch {
				ctx.Restore(before)
				break
			}
			main.Value.Functions = append(main.Value.Functions, function.Value)
			skipNewlines(ctx)
		}
		return Match[*ast.MainFunction]{Kind: Matched, Value: main.Value, Start: start.Pos, End: ctx.Pos}
	})
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
