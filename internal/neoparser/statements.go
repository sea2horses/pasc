package neoparser

import (
	"pseint-compiled/internal/ast"
	"pseint-compiled/internal/lexer"
	"pseint-compiled/internal/models"
	"pseint-compiled/internal/semantic"
)

func Statement() Pattern[ast.Stmt] {
	return OneOf(
		Declaration(),
		Assignment(),
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

func StatementBlock() Pattern[[]ast.Stmt] {
	return SepBy(
		Statement(),
		Tok(lexer.NEWLINE),
	)
}

func MainFunction() Pattern[*ast.MainFunction] {
	return Map(
		Seq2(
			Locate(
				Right(
					Kw(lexer.ALGORITMO),
					Name(),
				),
			),
			Left(
				StatementBlock(),
				Seq2(
					Newline(),
					Kw(lexer.FINALGORITMO),
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
						Name(),
						Tok(lexer.COMMA),
					),
					Right(
						Kw(lexer.COMO),
						Type(),
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
				Expression(),
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
				Name(),
				Between(
					Tok(lexer.L_BRACKET),
					SepBy1(
						Expression(),
						Tok(lexer.COMMA),
					),
					Tok(lexer.R_BRACKET),
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
				Expression(),
				TimeUnit(),
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
				Kw(lexer.PANTALLA),
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
