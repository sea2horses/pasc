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
			NodeInfo:     node.NodeInfo,
			Int:          node.Int,
			Frac:         node.Frac,
			ResolvedType: a.type_table.MakeInference(),
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
		if variable, ok := v.(*tast.VariableExpr); ok && variable.Symbol.Kind == semantic.SymbolFunction {
			if function, ok := semantic.Resolve(variable.Type()).(semantic.FunctionType); ok && len(function.Params) == 0 {
				call := &ast.Postfix{NodeInfo: node.NodeInfo, Target: node}
				return a.analyze_call(call, variable)
			}
		}
		return v
	case *ast.UnaryOperation:
		return a.analyze_unary_operation(node)
	case *ast.BinaryOperation:
		return a.analyze_binary_operation(node)
	case *ast.Postfix:
		return a.analyze_postfix(node)
	}

	diagnostics.Dbg("Could not find the expression type")
	diagnostics.Dbgfmt("%T\n", expr)
	a.Report(expr.NodeSpan(), "expression is not valid or implemented")
	return tast.ErrorExpr{NodeInfo: expr.Info()}
}

func (a *Analyzer) analyze_postfix(postfix *ast.Postfix) tast.TypedExpr {
	// The previous syntax allowed multidimensional indexing as a chain
	// (matrix[0][1]). Once postfix syntax is unified, those nodes are nested;
	// collapse a chain rooted at an array into one rank-wide indexing operation.
	chain := []*ast.Postfix{postfix}
	base := postfix.Target
	for {
		inner, ok := base.(*ast.Postfix)
		if !ok {
			break
		}
		chain = append(chain, inner)
		base = inner.Target
	}

	if len(chain) > 1 {
		baseExpr := a.analyze_postfix_target(base)
		baseType := semantic.Type(nil)
		if baseExpr != nil {
			baseType = semantic.Resolve(baseExpr.Type())
		}
		_, unresolved := baseType.(*semantic.InferType)
		if baseExpr != nil && (semantic.IsArray(baseType) || unresolved) {
			arguments := make([]ast.Expr, 0)
			for i := len(chain) - 1; i >= 0; i-- {
				arguments = append(arguments, chain[i].Arguments...)
			}
			combined := &ast.Postfix{
				NodeInfo:  postfix.NodeInfo,
				Target:    base,
				Arguments: arguments,
			}
			return a.analyze_indexing(combined, baseExpr)
		}
	}

	target := a.analyze_postfix_target(postfix.Target)
	if target == nil || semantic.IsInvalid(target.Type()) {
		return tast.ErrorExpr{NodeInfo: postfix.NodeInfo}
	}

	if semantic.IsArray(target.Type()) {
		return a.analyze_indexing(postfix, target)
	}
	if _, ok := semantic.Resolve(target.Type()).(*semantic.InferType); ok {
		return a.analyze_indexing(postfix, target)
	}
	if _, ok := semantic.Resolve(target.Type()).(semantic.FunctionType); ok {
		return a.analyze_call(postfix, target)
	}

	a.Report(postfix.Target.NodeSpan(), "expression of type '%s' is neither callable nor indexable", target.Type())
	return tast.ErrorExpr{NodeInfo: postfix.NodeInfo}
}

func (a *Analyzer) analyze_postfix_target(expr ast.Expr) tast.TypedExpr {
	if variable, ok := expr.(*ast.Variable); ok {
		if symbol, found := a.scope.Lookup(variable.Name); found && symbol.Kind == semantic.SymbolFunction {
			return &tast.VariableExpr{NodeInfo: variable.NodeInfo, Symbol: symbol}
		}
	}
	return a.analyze_expression(expr)
}

func (a *Analyzer) analyze_binary_operation(bo *ast.BinaryOperation) tast.TypedExpr {
	/* Analyze both inner expressions */
	LHS := a.analyze_expression(bo.LHS)
	RHS := a.analyze_expression(bo.RHS)
	/* Invalid by default */
	var resolved_type semantic.Type = semantic.Invalid_T{}

	ok := a.assert_nonvoid(LHS, RHS)
	if !ok {
		return tast.ErrorExpr{
			NodeInfo: bo.NodeInfo,
		}
	}

	if !semantic.IsInvalid(LHS.Type()) && !semantic.IsInvalid(RHS.Type()) {
		/* Check if we can perform the operation */
		rule, ok := semantic.ResolveBinaryOperation(bo.Op.Type, expressionTypeForOperation(LHS), expressionTypeForOperation(RHS))
		if ok {
			// Apply operand constraints first. Untyped literals resolve directly;
			// already typed operands may then need an explicit conversion.
			LHS, ok = coerce_to(LHS, rule.Left)
			if !ok {
				return tast.ErrorExpr{NodeInfo: bo.NodeInfo}
			}
			RHS, ok = coerce_to(RHS, rule.Right)
			if !ok {
				return tast.ErrorExpr{NodeInfo: bo.NodeInfo}
			}
			/* Check if we need to do a cast in LHS and RHS */
			if rule.LeftConversion != nil {
				LHS, ok = applyConversion(LHS, rule.LeftConversion)
				if !ok {
					return tast.ErrorExpr{NodeInfo: bo.NodeInfo}
				}
			}
			if rule.RightConversion != nil {
				RHS, ok = applyConversion(RHS, rule.RightConversion)
				if !ok {
					return tast.ErrorExpr{NodeInfo: bo.NodeInfo}
				}
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
	var resolved_type semantic.Type = semantic.Invalid_T{}

	ok := a.assert_nonvoid(typed_expr)
	if !ok {
		return tast.ErrorExpr{
			NodeInfo: uo.NodeInfo,
		}
	}

	if !semantic.IsInvalid(typed_expr.Type()) {
		/* Check if we can perform the operation */
		rule, ok := semantic.ResolveUnaryOperation(uo.Op.Type, expressionTypeForOperation(typed_expr))
		if ok {
			typed_expr, ok = coerce_to(typed_expr, rule.Operand)
			if !ok {
				return tast.ErrorExpr{NodeInfo: uo.NodeInfo}
			}
			/* Check if we got to do a cast */
			if rule.Conversion != nil {
				typed_expr, ok = applyConversion(typed_expr, rule.Conversion)
				if !ok {
					return tast.ErrorExpr{NodeInfo: uo.NodeInfo}
				}
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
