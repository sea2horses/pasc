package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/semantic"
)

func (a *Analyzer) analyze_while(while *ast.While) *tast.While {
	diagnostics.Dbg("Analyzing while...")
	/* Analyze the condition */
	typed_condition := a.analyze_expression(while.Condition)
	stmts := a.analyze_statement_block(while.Stmts)
	/* Assert the condition type to be boolean */
	ok := a.assert_type(semantic.Bool_T{}, typed_condition)
	if !ok {
		return nil
	}

	return &tast.While{
		NodeInfo:  while.NodeInfo,
		Condition: typed_condition,
		Stmts:     stmts,
	}
}

func (a *Analyzer) analyze_dowhile(dw *ast.DoWhile) *tast.DoWhile {
	/* Analyze the condition */
	typed_condition := a.analyze_expression(dw.Condition)
	stmts := a.analyze_statement_block(dw.Stmts)
	/* Assert the condition type to be boolean */
	ok := a.assert_type(semantic.Bool_T{}, typed_condition)
	if !ok {
		return nil
	}

	return &tast.DoWhile{
		NodeInfo:  dw.NodeInfo,
		Condition: typed_condition,
		Stmts:     stmts,
		Until:     dw.Until,
	}
}

/* For ONLY works with floats and numeric values!! */
func (a *Analyzer) analyze_for(f *ast.For) *tast.For {
	variable, ok := f.Start.Target.(*ast.Variable)
	if !ok {
		a.Report(f.Start.NodeSpan(), "expected variable")
		return nil
	}

	start := a.analyze_expression(f.Start.Content)
	end := a.analyze_expression(f.Until)

	targetType := semantic.Real_T{}

	valid := true
	if start == nil || end == nil {
		return nil
	}
	start, ok = coerce_to(start, targetType)
	if !ok || start == nil {
		got := semantic.Type(semantic.Invalid_T{})
		if start != nil {
			got = start.Type()
		}
		a.Report(f.Start.Content.NodeSpan(), "%s", ErrIncorrectType{Expected: targetType, Got: got})
		valid = false
	}
	end, ok = coerce_to(end, targetType)
	if !ok || end == nil {
		got := semantic.Type(semantic.Invalid_T{})
		if end != nil {
			got = end.Type()
		}
		a.Report(f.Until.NodeSpan(), "%s", ErrIncorrectType{Expected: targetType, Got: got})
		valid = false
	}

	var step tast.TypedExpr

	if f.Step != nil {
		step = a.analyze_expression(f.Step)
		if step == nil {
			return nil
		}
		step, ok = coerce_to(step, targetType)
		if !ok || step == nil {
			got := semantic.Type(semantic.Invalid_T{})
			if step != nil {
				got = step.Type()
			}
			a.Report(f.Step.NodeSpan(), "%s", ErrIncorrectType{Expected: targetType, Got: got})
			valid = false
		}
	}

	// The generated iterator yields float64, so the loop variable must have a
	// real type even when it shadows an outer integer variable.
	a.new_scope()
	symbol := &semantic.Symbol{
		Name:     variable.Name,
		Kind:     semantic.SymbolVariable,
		Type:     targetType,
		Declared: variable.NodeSpan(),
		Mutable:  true,
	}
	a.scope.Declare(symbol)
	typedVariable := &tast.VariableExpr{NodeInfo: variable.NodeInfo, Symbol: symbol}
	stmts := a.analyze_statement_block(f.Stmts)
	a.destroy_scope()
	if !valid {
		return nil
	}

	return &tast.For{
		Var:   typedVariable,
		Start: start,
		Until: end,
		Step:  step,
		Stmts: stmts,
	}
}

// analyze_foreach checks "Para Cada x De v": v must be an array, and x is a
// new reference to each element, so assigning to x writes into v.
func (a *Analyzer) analyze_foreach(f *ast.ForEach) *tast.ForEach {
	array := a.analyze_expression(f.Array)
	if array == nil || semantic.IsInvalid(array.Type()) {
		return nil
	}
	arrayType, ok := semantic.Resolve(array.Type()).(semantic.ArrayType)
	if !ok {
		a.Report(f.Array.NodeSpan(), "'Para Cada' needs an array, got '%s'", array.Type())
		return nil
	}

	a.new_scope()
	symbol := &semantic.Symbol{
		Name:        f.Var.Name,
		Kind:        semantic.SymbolVariable,
		Type:        arrayType.Elem,
		Declared:    f.Var.NodeSpan(),
		Mutable:     true,
		ByReference: true,
	}
	a.scope.Declare(symbol)
	stmts := a.analyze_statement_block(f.Stmts)
	a.destroy_scope()

	return &tast.ForEach{
		NodeInfo: f.NodeInfo,
		Var:      &tast.VariableExpr{NodeInfo: f.Var.NodeInfo, Symbol: symbol},
		Array:    array,
		Stmts:    stmts,
	}
}
