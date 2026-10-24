package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/semantic"
)

func (a *Analyzer) analyze_expression(expr ast.Expr) tast.TypedExpr {
	diagnostics.Dbg("Analyzing expression")
	switch node := expr.(type) {
	case *ast.NumberLiteral:
		return &tast.NumberLiteral{
			NodeInfo: node.NodeInfo,
			Int:      node.Int,
			Frac:     node.Frac,
		}
	case *ast.StringLiteral:
		return &tast.StringLiteral{
			NodeInfo: node.NodeInfo,
			Value:    node.Content,
		}
	case *ast.BoolLiteral:
		return &tast.BooleanLiteral{
			NodeInfo: node.NodeInfo,
			Value:    node.Value,
		}
	case *ast.Variable:
		v := a.analyze_variable(node)
		return v
	case *ast.UnaryOperation:
		return a.analyze_unary_operation(node)
	case *ast.BinaryOperation:
		return a.analyze_binary_operation(node)
	case *ast.Index:
		return a.analyze_indexing(node)
	case *ast.Call:
		return a.analyze_call(node)
	}

	diagnostics.Dbg("Could not find the expression type")
	diagnostics.Dbgfmt("%T\n", expr)
	a.Report(expr.NodeSpan(), "expression is not valid or implemented")
	return tast.ErrorExpr{NodeInfo: expr.Info()}
}

func (a *Analyzer) analyze_binary_operation(bo *ast.BinaryOperation) tast.TypedExpr {
	/* Analyze both inner expressions */
	LHS := a.analyze_expression(bo.LHS)
	RHS := a.analyze_expression(bo.RHS)
	/* Invalid by default */
	var resolved_type semantic.Type = semantic.InvalidType{}

	ok := a.assert_nonvoid(LHS, RHS)
	if !ok {
		return tast.ErrorExpr{
			NodeInfo: bo.NodeInfo,
		}
	}

	if !semantic.IsInvalid(LHS.Type()) && !semantic.IsInvalid(RHS.Type()) {
		/* Check if we can perform the operation */
		rule, ok := semantic.ResolveBinaryOperation(bo.Op.Type, LHS.Type(), RHS.Type())
		if ok {
			/* Check if we need to do a cast in LHS and RHS */
			if rule.LeftConversion != semantic.ConversionNone {
				LHS = applyConversion(LHS, rule.LeftConversion)
			}
			if rule.RightConversion != semantic.ConversionNone {
				RHS = applyConversion(RHS, rule.RightConversion)
			}
			resolved_type = rule.Result
		} else {
			a.Report(bo.NodeInfo.Span, "%s", ErrBiOperationNotAllowed{LHS: LHS.Type(), RHS: RHS.Type(), Operation: bo.Op.Type}.Error())
			return tast.ErrorExpr{
				NodeInfo: bo.NodeInfo,
			}
		}
	}

	return &tast.BinaryOperation{
		NodeInfo:     bo.NodeInfo,
		LHS:          LHS,
		RHS:          RHS,
		Op:           a.analyze_operator(bo.Op),
		ResolvedType: resolved_type,
	}
}

func (a *Analyzer) analyze_unary_operation(uo *ast.UnaryOperation) tast.TypedExpr {
	/* Analyze the inner expression */
	typed_expr := a.analyze_expression(uo.Expr)
	/* Invalid by default */
	var resolved_type semantic.Type = semantic.InvalidType{}

	ok := a.assert_nonvoid(typed_expr)
	if !ok {
		return tast.ErrorExpr{
			NodeInfo: uo.NodeInfo,
		}
	}

	if !semantic.IsInvalid(typed_expr.Type()) {
		/* Check if we can perform the operation */
		rule, ok := semantic.ResolveUnaryOperation(uo.Op.Type, typed_expr.Type())
		if ok {
			/* Check if we got to do a cast */
			if rule.Conversion != semantic.ConversionNone {
				typed_expr = applyConversion(typed_expr, rule.Conversion)
			}
			/* First attach the resolved type */
			resolved_type = rule.Result
		} else {
			a.Report(uo.NodeInfo.Span, "%s", ErrUnOperationNotAllowed{Type: typed_expr.Type(), Operation: uo.Op.Type}.Error())
			return tast.ErrorExpr{
				NodeInfo: uo.NodeInfo,
			}
		}
	}

	return &tast.UnaryOperation{
		NodeInfo:     uo.NodeInfo,
		Expr:         typed_expr,
		Op:           a.analyze_operator(uo.Op),
		ResolvedType: resolved_type,
	}
}
