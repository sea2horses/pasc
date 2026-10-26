package parser

import (
	"pascc/internal/ast"
	"pascc/internal/lexer"
	"pascc/internal/models"
	"pascc/internal/semantic"
	"slices"
)

func Operand() Pattern[ast.Expr] {
	return Map(
		Bind(Many(UnaryOperator()), func(ops []*ast.Operator) Pattern[Pair[[]*ast.Operator, ast.Expr]] {
			base := ChainBind(Atom(), Postfix)
			if len(ops) > 0 {
				base = Expected(base, ExpectedExpression)
			}
			return Map(base, func(expr ast.Expr) Pair[[]*ast.Operator, ast.Expr] {
				return Pair[[]*ast.Operator, ast.Expr]{First: ops, Second: expr}
			})
		}),
		func(p Pair[[]*ast.Operator, ast.Expr]) ast.Expr {
			new_expr := p.Second
			slices.Reverse(p.First)
			for _, o := range p.First {
				new_expr = &ast.UnaryOperation{
					NodeInfo: ast.NodeInfo{
						Span: models.JoinSpans(new_expr.NodeSpan(), o.Span),
					},
					Op:   *o,
					Expr: new_expr,
				}
			}
			return new_expr
		},
	)
}

func Postfix(target ast.Expr) Pattern[ast.Expr] {
	return MapWithLocation(
		OneOf(
			Between(
				Tok(lexer.L_PARENTHESES),
				SepBy(Expression(), Tok(lexer.COMMA)),
				Expected(Tok(lexer.R_PARENTHESES), ExpectedToken{Kind: lexer.R_PARENTHESES}),
			),
			Between(
				Tok(lexer.L_BRACKET),
				SepBy(Expression(), Tok(lexer.COMMA)),
				Expected(Tok(lexer.R_BRACKET), ExpectedToken{Kind: lexer.R_BRACKET}),
			),
		),
		func(arguments []ast.Expr, span models.Span) ast.Expr {
			return &ast.Postfix{
				NodeInfo: ast.NodeInfo{
					Span: models.JoinSpans(target.NodeSpan(), span),
				},
				Arguments: arguments,
				Target:    target,
			}
		},
	)
}

func Expression() Pattern[ast.Expr] {
	return PatternFunc[ast.Expr](func(ctx *Context) Match[ast.Expr] {
		operandStack := []ast.Expr{}
		operatorStack := []ast.Operator{}

		start := ctx.Pos

		execute := func() {
			n := len(operandStack)

			lhs := operandStack[n-2]
			rhs := operandStack[n-1]
			operandStack = operandStack[:n-2]

			n = len(operatorStack)
			op := operatorStack[n-1]
			operatorStack = operatorStack[:n-1]

			expr := &ast.BinaryOperation{
				NodeInfo: ast.NodeInfo{
					Span: models.JoinSpans(
						lhs.NodeSpan(),
						rhs.NodeSpan(),
						op.NodeSpan(),
					),
				},
				LHS: lhs,
				Op:  op,
				RHS: rhs,
			}

			operandStack = append(operandStack, expr)
		}

		// Parse the first operand
		first := Operand().Match(ctx)

		if !first.OK() {
			return Match[ast.Expr]{
				Kind:  first.Kind,
				Start: start,
				End:   first.End,
				Err:   first.Err,
			}
		}

		operandStack = append(operandStack, first.Value)

		for {
			operator := BinaryOperator().Match(ctx)

			if !operator.OK() {

				if operator.Kind == Failed {
					return Match[ast.Expr]{
						Kind:  Failed,
						Start: operator.Start,
						End:   operator.End,
						Err:   operator.Err,
					}
				}
				break
			}

			current := operator.Value
			currentPrec := current.Precedence()

			// Reduce pending operators
			for len(operatorStack) > 0 {
				top := operatorStack[len(operatorStack)-1]
				topPrec := top.Precedence()

				// Assuming all binary operators are left-associative
				shouldReduce := topPrec > currentPrec ||
					(topPrec == currentPrec &&
						current.Type.Associativity() == semantic.LeftAssociative)

				if !shouldReduce {
					break
				}

				execute()
			}

			operatorStack = append(operatorStack, *current)

			// Once an operator matches, its RHS is mandatory
			operand := Expected(Operand(), ExpectedExpression).Match(ctx)

			if !operand.OK() {
				return Match[ast.Expr]{
					Kind:  Failed,
					Start: start,
					End:   ctx.Pos,
					Err:   operand.Err,
				}
			}

			operandStack = append(operandStack, operand.Value)
		}

		// Reduce remaining operators
		for len(operatorStack) > 0 {
			execute()
		}

		return Match[ast.Expr]{
			Kind:  Matched,
			Value: operandStack[0],
			Start: start,
			End:   ctx.Pos,
		}
	})
}
