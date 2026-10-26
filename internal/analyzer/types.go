package analyzer

import (
	"strings"

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
	return a.assert_nontype(semantic.Void_T{}, ErrUnexpectedVoid, exprs...)
}

func unify(target semantic.Type, source semantic.Type) bool {
	target = semantic.Resolve(target)
	source = semantic.Resolve(source)
	if targetInfer, ok := target.(*semantic.InferType); ok {
		if sourceInfer, sourceOK := source.(*semantic.InferType); sourceOK && targetInfer == sourceInfer {
			return true
		}
	}

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
	if number, ok := expr.(*tast.NumberLiteral); ok {
		// Numeric literals carry an inference variable until context selects a
		// concrete type. Resolve that variable directly, without a Cast node.
		if infer, ok := target.(*semantic.InferType); ok && infer.Resolved == nil {
			// Match the normal untyped-number default: integer spelling becomes
			// integer, while a decimal spelling becomes real.
			if number.Frac == "" {
				number.ResolvedType = semantic.Integer_T{}
			} else {
				number.ResolvedType = semantic.Real_T{}
			}
			infer.Resolved = number.ResolvedType
			return expr, true
		}
		if !semantic.IsNumeric(target) {
			return nil, false
		}
		if semantic.IsInteger(target) && strings.Trim(number.Frac, "0") != "" {
			return nil, false
		}
		if infer, ok := number.ResolvedType.(*semantic.InferType); ok && infer.Resolved == nil {
			infer.Resolved = target
			return expr, true
		}
		if number.ResolvedType == nil || semantic.EqualTypes(number.Type(), semantic.UntypedNumber_T{}) {
			number.ResolvedType = target
			return expr, true
		}
	}
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

func expressionTypeForOperation(expr tast.TypedExpr) semantic.Type {
	if number, ok := expr.(*tast.NumberLiteral); ok && number.Frac != "" {
		return semantic.Real_T{}
	}
	return expr.Type()
}

// defaultUntypedNumber settles a literal when no surrounding expression can
// provide a type. Whole-number spelling defaults to integer; decimal spelling
// defaults to real.
func defaultUntypedNumber(expr tast.TypedExpr) tast.TypedExpr {
	number, ok := expr.(*tast.NumberLiteral)
	if !ok || !semantic.EqualTypes(number.Type(), semantic.UntypedNumber_T{}) {
		return expr
	}
	target := semantic.Type(semantic.Integer_T{})
	if number.Frac != "" {
		target = semantic.Real_T{}
	}
	_, _ = coerce_to(number, target)
	return number
}

func isUnresolvedInference(t semantic.Type) bool {
	infer, ok := semantic.Resolve(t).(*semantic.InferType)
	return ok && infer.Resolved == nil
}
