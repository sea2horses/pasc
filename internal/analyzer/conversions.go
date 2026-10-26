package analyzer

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
)

func tryConvert(src tast.TypedExpr, to semantic.Type) (tast.TypedExpr, bool) {
	from := semantic.Resolve(src.Type())
	to = semantic.Resolve(to)
	if semantic.EqualTypes(from, to) {
		return src, true
	}
	plan := semantic.CanCast(from, to)
	if plan == nil {
		return nil, false
	}
	return applyConversion(src, plan)
}

// applyConversion applies a conversion plan already selected by semantic
// resolution. Keeping the source check here catches stale or mismatched plans.
func applyConversion(operand tast.TypedExpr, plan *semantic.ImplicitCast) (tast.TypedExpr, bool) {
	if plan == nil {
		return operand, true
	}
	if !semantic.EqualTypes(semantic.Resolve(operand.Type()), semantic.Resolve(plan.From)) {
		return nil, false
	}
	kind := plan.ConversionKind
	return &tast.Cast{
		TargetType:     plan.To,
		ConversionKind: &kind,
		Expr:           operand,
	}, true
}
