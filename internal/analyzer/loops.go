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
	ok := a.assert_type(semantic.BooleanType{}, typed_condition)
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
	ok := a.assert_type(semantic.BooleanType{}, typed_condition)
	if !ok {
		return nil
	}

	return &tast.DoWhile{
		NodeInfo:  dw.NodeInfo,
		Condition: typed_condition,
		Stmts:     stmts,
	}
}

/* For ONLY works with floats and numeric values!! */
func (a *Analyzer) analyze_for(f *ast.For) *tast.For {
	_ = a.analyze_statement(f.Start)

	exp := a.analyze_expression(f.Start.Target)
	variable, ok := exp.(*tast.VariableExpr)
	if !ok {
		a.Report(f.Start.NodeSpan(), "expected variable")
	}

	start := a.analyze_expression(f.Start.Content)
	end := a.analyze_expression(f.Until)

	targetType := semantic.RealType{}

	start, ok = coerce_to(start, targetType)
	if !ok {
		a.Report(start.NodeSpan(), "%s", ErrIncorrectType{Expected: targetType, Got: start.Type()})
	}
	end, ok = coerce_to(end, targetType)
	if !ok {
		a.Report(end.NodeSpan(), "%s", ErrIncorrectType{Expected: targetType, Got: end.Type()})
	}

	var step tast.TypedExpr

	if f.Step != nil {
		step = a.analyze_expression(f.Step)
		step, ok = coerce_to(step, targetType)
		if !ok {
			a.Report(end.NodeSpan(), "%s", ErrIncorrectType{Expected: targetType, Got: step.Type()})
		}
	}

	/* ALL must coerce to float */

	return &tast.For{
		Var:   variable,
		Start: start,
		Until: end,
		Step:  step,
		Stmts: a.analyze_statement_block(f.Stmts),
	}
}
