package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/semantic"
	"reflect"
)

func (a *Analyzer) analyze_variable(variable *ast.Variable) tast.TypedExpr {
	diagnostics.Dbg("Analyzing variable: ", variable.Name)
	/* Get variable symbol */
	symbol, ok := a.scope.Lookup(variable.Name)

	if !ok || symbol == nil {
		diagnostics.Dbg("Not found!")
		a.Report(variable.NodeSpan(), "%s", ErrVariableDoesntExist{Name: variable.Name}.Error())
		return &tast.ErrorExpr{NodeInfo: variable.NodeInfo}
	}

	diagnostics.Dbg("Found! Symbol: ", symbol)
	return &tast.VariableExpr{
		NodeInfo: variable.NodeInfo,
		Symbol:   symbol,
	}
}

func (a *Analyzer) analyze_single_decl(decl *ast.Declaration, name string, t semantic.Type) *semantic.Symbol {
	if _, exists := a.scope.Lookup(name); exists {
		a.Report(decl.Span, "%s", ErrSymbolAlreadyExists{Name: name}.Error())
		return nil
	}

	symbol := &semantic.Symbol{
		Name:     name,
		Kind:     semantic.SymbolVariable,
		Type:     t,
		Declared: decl.Span,
		Mutable:  true,
	}

	a.scope.Declare(symbol)
	return symbol
}

func (a *Analyzer) analyze_declaration(decl *ast.Declaration) *tast.Declaration {
	diagnostics.Dbg("Analyzing declaration")
	ok := true
	/* Let's declare this bitch */
	/* Check the type exists! */
	t := a.analyze_type_ref(decl.Type)

	if t == nil {
		a.Report(decl.Type.Span, "%s", "unknown type")
		ok = false
	}

	symbols := []*semantic.Symbol{}
	for _, n := range decl.Names {
		symbol := a.analyze_single_decl(decl, n, t)

		if symbol == nil {
			ok = false
		}

		symbols = append(symbols, symbol)
	}

	if !ok {
		return nil
	}

	return &tast.Declaration{
		NodeInfo: decl.NodeInfo,
		Symbols:  symbols,
		Type:     t,
	}
}

/* TODO: 'strict' assignment rule where implicit assignment is forbidden */
func (a *Analyzer) analyze_assignment(ass *ast.Assignment) *tast.Assignment {
	diagnostics.Dbg("Analyzing assignment")
	/* Check if LHS of assignment is a variable node only */
	variable, ok := ass.Target.(*ast.Variable)
	typed_expr := a.analyze_expression(ass.Content)
	nonvoid := a.assert_nonvoid(typed_expr)
	declarative := false

	diagnostics.Dbg("Target: ", ass.Target, reflect.TypeOf(ass.Target))

	if !nonvoid {
		diagnostics.Dbg("Void expression detected")
		return nil
	}

	diagnostics.Dbg("Passing through...")
	/* Let's check it out */
	if ok {
		diagnostics.Dbg("Direct variable assignment detected")
		if _, exists := a.scope.Lookup(variable.Name); !exists {
			diagnostics.Dbg("Making implicit declaration")
			typed_expr = defaultUntypedNumber(typed_expr)
			/* Implicit declaration */
			a.scope.Declare(
				&semantic.Symbol{
					Name:     variable.Name,
					Kind:     semantic.SymbolVariable,
					Type:     typed_expr.Type(),
					Declared: ass.NodeSpan(),
					Mutable:  true,
				},
			)

			declarative = true
		}
	}
	target := a.analyze_expression(ass.Target)
	diagnostics.Dbg("Target <typed>: ", target, reflect.TypeOf(target))

	lvalue, ok := target.(tast.LValue)
	if !ok {
		a.Report(ass.Target.NodeSpan(), ErrLValueError)
		return nil
	}

	if !lvalue.IsMutable() {
		a.Report(ass.NodeSpan(), ErrImmutable)
		a.Info(lvalue.AssignmentOrigin().Span, "%s", lvalue.AssignmentOrigin().Message)
		return nil
	}

	/* Coerce to the lvalue type, allowing contextual literal typing. */
	originalType := typed_expr.Type()
	typed_expr, ok = coerce_to(typed_expr, target.Type())
	if !ok {
		a.Report(ass.NodeSpan(), "%s", ErrIncorrectType{Expected: target.Type(), Got: originalType}.Error())
		a.Info(lvalue.AssignmentOrigin().Span, "%s", lvalue.AssignmentOrigin().Message)
		return nil
	}

	return &tast.Assignment{
		NodeInfo:    ass.NodeInfo,
		Target:      lvalue,
		Value:       typed_expr,
		Declarative: declarative,
	}
}
