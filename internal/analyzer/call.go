package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/i18n"
	"pascc/internal/semantic"
)

func (a *Analyzer) analyze_call(call *ast.Postfix, callable tast.TypedExpr) tast.TypedExpr {

	fnType, ok := semantic.Resolve(callable.Type()).(semantic.FunctionType)
	if !ok {
		a.errorf(call.Target.NodeSpan(), i18n.NotCallable, semantic.DisplayType(callable.Type()))

		return tast.ErrorExpr{
			NodeInfo: call.NodeInfo,
		}
	}

	if len(call.Arguments) != len(fnType.Params) {
		a.errorf(call.NodeSpan(), i18n.ArgCountMismatch, len(fnType.Params), len(call.Arguments))

		return tast.ErrorExpr{
			NodeInfo: call.NodeInfo,
		}
	}

	args := make([]tast.TypedExpr, 0, len(call.Arguments))
	byReference := append([]bool(nil), fnType.ByReference...)

	for i, arg := range call.Arguments {
		typedArg := a.analyze_expression(arg)
		got := typedArg.Type()
		isReference := i < len(byReference) && byReference[i]
		if isReference {
			lvalue, ok := typedArg.(tast.LValue)
			if !ok || !lvalue.IsMutable() {
				a.errorf(arg.NodeSpan(), i18n.ArgMustBeVariable, i+1)
				return tast.ErrorExpr{NodeInfo: call.NodeInfo}
			}
			if !unify(fnType.Params[i], typedArg.Type()) {
				a.errorf(arg.NodeSpan(), i18n.ArgTypeMismatch, i+1, semantic.DisplayType(fnType.Params[i]), semantic.DisplayType(got))
				return tast.ErrorExpr{NodeInfo: call.NodeInfo}
			}
			args = append(args, typedArg)
			continue
		}

		// This is EXACTLY the same target typing as assignment.
		typedArg, ok = coerce_to(typedArg, fnType.Params[i])
		if !ok {
			a.errorf(arg.NodeSpan(), i18n.ArgTypeMismatch, i+1, semantic.DisplayType(fnType.Params[i]), semantic.DisplayType(got))

			return tast.ErrorExpr{
				NodeInfo: call.NodeInfo,
			}
		}

		args = append(args, typedArg)
	}

	return &tast.Call{
		NodeInfo:    call.NodeInfo,
		Callable:    callable,
		Arguments:   args,
		ByReference: byReference,
		Return:      fnType.Return,
	}
}
