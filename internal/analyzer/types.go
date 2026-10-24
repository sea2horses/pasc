package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
)

func (a *Analyzer) analyze_type_ref(typeref *ast.TypeRef) semantic.Type {
	return a.type_table.Get(typeref.Name)
}

func (a *Analyzer) assert_type(target semantic.Type, exprs ...tast.TypedExpr) bool {
	ok := true
	for _, expr := range exprs {
		if expr.Type() != target {
			a.Report(expr.NodeSpan(), "%s", ErrExpectedType{Type: target}.Error())
			ok = false
		}
	}
	return ok
}

func (a *Analyzer) assert_nontype(target semantic.Type, msg string, exprs ...tast.TypedExpr) bool {
	ok := true
	for _, expr := range exprs {
		if expr.Type() == target {
			a.Report(expr.NodeSpan(), "%s", msg)
			ok = false
		}
	}
	return ok
}

func (a *Analyzer) assert_nonvoid(exprs ...tast.TypedExpr) bool {
	return a.assert_nontype(semantic.VoidType{}, ErrUnexpectedVoid, exprs...)
}

func unify(target semantic.Type, source semantic.Type) bool {
	target = semantic.Resolve(target)
	source = semantic.Resolve(source)

	// Target is unknown: learn from source.
	if infer, ok := target.(*semantic.InferType); ok && infer.Resolved == nil {
		infer.Resolved = source
		return true
	}

	// Source is unknown: propagate the constraint.
	if infer, ok := source.(*semantic.InferType); ok && infer.Resolved == nil {
		infer.Resolved = target
		return true
	}

	return semantic.EqualTypes(target, source)
}

func coerce_to(expr tast.TypedExpr, target semantic.Type) (tast.TypedExpr, bool) {
	target = semantic.Resolve(target)
	source := semantic.Resolve(expr.Type())

	/* Unify in case */
	ok := unify(target, source)
	/* Try conversion */
	if ok {
		return expr, true
	}

	converted, ok := tryConvert(expr, target)
	if ok {
		return converted, true
	}

	return nil, false
}
