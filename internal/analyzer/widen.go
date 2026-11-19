package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/semantic"
)

// PSeInt does not distinguish integers from reals in undeclared variables: an
// accumulator that starts at 0 may later hold 2.5. pascc types an implicit
// variable from its first assignment, so when a later assignment stores a
// real into an implicit integer (an implicit real -> integer conversion), the
// variable is widened to real and the program is analyzed again. Widening one
// variable can make others real, so this repeats until nothing changes.
const maxWideningPasses = 8

func (a *Analyzer) Analyze(program *ast.MainFunction) (*tast.MainFunction, []diagnostics.Diagnostic) {
	a.widened = map[string]bool{}
	for pass := 1; ; pass++ {
		main, ds := a.analyzeOnce(program)
		if pass == maxWideningPasses || a.Checkpoint().AnyError() {
			return main, ds
		}
		widened := a.collectWidening(main)
		if len(widened) == 0 {
			return main, ds
		}
		for _, key := range widened {
			diagnostics.Dbgfmt("widening implicit variable %s to real\n", key)
			a.widened[key] = true
		}
	}
}

func (a *Analyzer) implicitKey(name string) string {
	return a.currentFunction + "/" + semantic.NormalizeName(name)
}

func (a *Analyzer) markImplicit(symbol *semantic.Symbol) {
	a.implicitKeys[symbol] = a.implicitKey(symbol.Name)
}

// widenedType replaces the type of an implicit variable or array element with
// real once an earlier pass found it storing reals.
func (a *Analyzer) widenedType(name string, t semantic.Type) semantic.Type {
	if a.widened[a.implicitKey(name)] {
		return semantic.Real_T{}
	}
	return t
}

// collectWidening finds implicit variables (and implicitly typed array
// elements) that are assigned a real through a real -> integer conversion.
func (a *Analyzer) collectWidening(main *tast.MainFunction) []string {
	var found []string
	seen := map[string]bool{}
	visit := func(stmt tast.Stmt) {
		assignment, ok := stmt.(*tast.Assignment)
		if !ok || assignment == nil {
			return
		}
		cast, ok := assignment.Value.(*tast.Cast)
		if !ok || cast.ConversionKind == nil || *cast.ConversionKind != semantic.ConversionRealToIntegerExact {
			return
		}
		target := tast.TypedExpr(assignment.Target)
		if index, ok := target.(*tast.Index); ok {
			target = index.Target
		}
		variable, ok := target.(*tast.VariableExpr)
		if !ok {
			return
		}
		key, implicit := a.implicitKeys[variable.Symbol]
		if !implicit || a.widened[key] || seen[key] {
			return
		}
		seen[key] = true
		found = append(found, key)
	}
	walkStmts(main.Stmts, visit, func(tast.TypedExpr) {})
	for _, function := range main.Functions {
		walkStmts(function.Stmts, visit, func(tast.TypedExpr) {})
	}
	return found
}
