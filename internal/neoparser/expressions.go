package neoparser

import (
	"pseint-compiled/internal/ast"
	"pseint-compiled/internal/lexer"
	"pseint-compiled/internal/models"
	"slices"
)

func Operand() Pattern[ast.Expr] {
	return Map(
		Seq2(
			Many(
				UnaryOperator(),
			),
			/* Allows postfix chaining hello[1][2]()[3] */
			ChainBind(
				Atom(),
				Postfix,
			),
		),
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
	return OneOf(
		Index(target),
		Call(target),
	)
}

func Index(target ast.Expr) Pattern[ast.Expr] {
	return MapWithLocation(
		Many1(
			Between(
				Tok(lexer.L_BRACKET),
				Expression(),
				Tok(lexer.R_BRACKET),
			),
		),
		func(indices []ast.Expr, span models.Span) ast.Expr {
			return &ast.Index{
				NodeInfo: ast.NodeInfo{
					Span: models.JoinSpans(target.NodeSpan(), span),
				},
				Indexes: indices,
				Target:  target,
			}
		},
	)
}

func Call(target ast.Expr) Pattern[ast.Expr] {
	return MapWithLocation(
		Between(
			Tok(lexer.L_PARENTHESES),
			SepBy(Expression(), Tok(lexer.COMMA)),
			Tok(lexer.R_PARENTHESES),
		),
		func(arguments []ast.Expr, span models.Span) ast.Expr {
			return &ast.Call{
				NodeInfo: ast.NodeInfo{
					Span: models.JoinSpans(target.NodeSpan(), span),
				},
				Callable:  target,
				Arguments: arguments,
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
				Kind:  Failed,
				Start: start,
				End:   ctx.Pos,
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
				if topPrec < currentPrec {
					break
				}

				execute()
			}

			operatorStack = append(operatorStack, *current)

			// Once an operator matches, its RHS is mandatory
			operand := Operand().Match(ctx)

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
