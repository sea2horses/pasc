package parser

import (
	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/models"
	"pascc/internal/semantic"
	"slices"
)

func (p *Parser) read_infix_to_postfix() ([]ast.Node, error) {
	operator_stack := []ast.Operator{}
	postfix_stack := []ast.Node{}

	for {
		/* Parse operand */
		expr, err := p.parse_operand()
		if err != nil {
			p.Report(p.currentSpan(), "expected operand")
			return nil, err
		}

		/* Add it directly to the postfix stack */
		postfix_stack = append(postfix_stack, expr)

		/* Grab operator */
		op := p.parse_operator()
		/* If there are no more operators, break */
		if op == nil {
			break
		}

		/* Else, if the current operator is lesser or equal than the one on top, we pop the stack */
		for len(operator_stack) > 0 &&
			semantic.BinaryOperatorPrecedence(operator_stack[len(operator_stack)-1].Type) >= semantic.BinaryOperatorPrecedence(op.Type) {
			/* Append top operator to postfix stack */
			postfix_stack = append(postfix_stack, operator_stack[len(operator_stack)-1])
			/* Chop last. element */
			operator_stack = operator_stack[0 : len(operator_stack)-1]
		}

		operator_stack = append(operator_stack, *op)
	}

	/* Unload the remaining operators */
	for len(operator_stack) > 0 {
		/* Append top operator to postfix stack */
		postfix_stack = append(postfix_stack, operator_stack[len(operator_stack)-1])
		/* Chop last. element */
		operator_stack = operator_stack[0 : len(operator_stack)-1]
	}

	return postfix_stack, nil
}

func (p *Parser) postfix_to_expression(chain []ast.Node) (ast.Expr, error) {
	for {
		diagnostics.Dbg("Postfix Iteration, expr: ", chain)

		if len(chain) == 1 {
			expr, ok := chain[0].(ast.Expr)
			if !ok {
				return nil, ErrInvalidExpression
			}

			diagnostics.Dbg("Expression made: ", expr)
			return expr, nil
		}

		swapped := false

		for i := 0; i+2 < len(chain); i++ {
			first, ok := chain[i].(ast.Expr)
			if !ok {
				continue
			}

			second, ok := chain[i+1].(ast.Expr)
			if !ok {
				continue
			}

			third, ok := chain[i+2].(ast.Operator)
			if !ok {
				continue
			}

			expr := &ast.BinaryOperation{
				LHS: first,
				RHS: second,
				Op:  third,
				NodeInfo: ast.NodeInfo{
					Span: models.JoinSpans(
						first.NodeSpan(),
						second.NodeSpan(),
						third.NodeSpan(),
					),
				},
			}

			chain = slices.Delete(chain, i, i+3)
			chain = slices.Insert(chain, i, ast.Node(expr))

			swapped = true

			// IMPORTANT:
			// chain changed, restart the scan.
			break
		}

		if !swapped {
			return nil, ErrInvalidExpression
		}
	}
}
