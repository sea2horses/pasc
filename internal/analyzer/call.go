package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
)

func (a *Analyzer) analyze_call(call *ast.Postfix, callable tast.TypedExpr) tast.TypedExpr {

	fnType, ok := semantic.Resolve(callable.Type()).(semantic.FunctionType)
	if !ok {
		a.Report(
			call.Target.NodeSpan(),
			"expression of type '%s' is not callable",
			callable.Type(),
		)

		return tast.ErrorExpr{
			NodeInfo: call.NodeInfo,
		}
	}

	if len(call.Arguments) != len(fnType.Params) {
		a.Report(
			call.NodeSpan(),
			"expected %d arguments, got %d",
			len(fnType.Params),
			len(call.Arguments),
		)

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
				a.Report(arg.NodeSpan(), "argument %d must be a mutable variable for a Por Referencia parameter", i+1)
				return tast.ErrorExpr{NodeInfo: call.NodeInfo}
			}
			if !unify(fnType.Params[i], typedArg.Type()) {
				a.Report(arg.NodeSpan(), "argument %d: expected '%s', got '%s'", i+1, fnType.Params[i], got)
				return tast.ErrorExpr{NodeInfo: call.NodeInfo}
			}
			args = append(args, typedArg)
			continue
		}

		// This is EXACTLY the same target typing as assignment.
		typedArg, ok = coerce_to(typedArg, fnType.Params[i])
		if !ok {
			a.Report(
				arg.NodeSpan(),
				"argument %d: expected '%s', got '%s'",
				i+1,
				fnType.Params[i],
				got,
			)

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
		Return:      semantic.Resolve(fnType.Return),
	}
}
