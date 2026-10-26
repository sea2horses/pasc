package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/semantic"
)

func (a *Analyzer) analyze_if(i *ast.If) *tast.If {
	diagnostics.Dbg("Analyzing if...")
	/* Turn the expression into a typed expression */
	typed_condition := a.analyze_expression(i.Condition)
	stmts := a.analyze_statement_block(i.Stmts)
	/* Assert the condition type to be boolean */
	ok := a.assert_type(semantic.Bool_T{}, typed_condition)
	if !ok {
		return nil
	}

	var else_branch *tast.Else = nil

	if i.Else != nil {
		else_branch = a.analyze_else(i.Else)
	}

	return &tast.If{
		NodeInfo:  i.NodeInfo,
		Condition: typed_condition,
		Stmts:     stmts,
		Else:      else_branch,
	}
}

func (a *Analyzer) analyze_else(e *ast.Else) *tast.Else {
	diagnostics.Dbg("Analyzing else...")
	stmts := a.analyze_statement_block(e.Stmts)
	return &tast.Else{NodeInfo: e.NodeInfo, Stmts: stmts}
}

func (a *Analyzer) analyze_switch(s *ast.Switch) *tast.Switch {
	diagnostics.Dbg("Analyzing switch...")

	typed_base := a.analyze_expression(s.Base)
	if semantic.IsInvalid(typed_base.Type()) {
		diagnostics.Dbg("Invalid base type")
	}

	cases := make([]*tast.Case, 0, len(s.Cases))
	for _, c := range s.Cases {
		typed_value := a.analyze_expression(c.Clause)
		/* Check that type is comparable */
		if !semantic.EqualTypes(typed_base.Type(), typed_value.Type()) {
			var ok bool
			typed_value, ok = tryConvert(typed_value, typed_base.Type())
			if !ok {
				a.Report(typed_value.NodeSpan(), "expression of type %s cannot be assigned to %s", typed_value.Type(), typed_base.Type())
				continue
			}
		}

		typed_stmts := a.analyze_statement_block(c.Stmts)
		cases = append(cases, &tast.Case{
			NodeInfo: c.NodeInfo,
			Value:    typed_value,
			Stmts:    typed_stmts,
		})
	}

	/* Parse default */
	default_stmts := make([]tast.Stmt, 0, len(s.Default))
	for _, d := range s.Default {
		default_stmts = append(default_stmts, a.analyze_statement(d)...)
	}

	return &tast.Switch{
		NodeInfo: s.NodeInfo,
		Base:     typed_base,
		Cases:    cases,
		Default:  default_stmts,
	}
}
