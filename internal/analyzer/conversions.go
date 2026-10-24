package analyzer

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
)

func tryConvert(src tast.TypedExpr, to semantic.Type) (*tast.Cast, bool) {
	conversion := semantic.CanCast(src.Type(), to)
	if conversion == nil {
		return nil, false
	}
	cast := applyConversion(src, *conversion)
	return cast, cast != nil
}

func applyConversion(operand tast.TypedExpr, conversion semantic.ConversionKind) *tast.Cast {
	lookup := semantic.SearchConversion(operand.Type(), conversion)
	if lookup == nil {
		return nil
	}
	return &tast.Cast{
		TargetType:     lookup.To,
		ConversionKind: &lookup.ConversionKind,
		Expr:           operand,
	}
}
