package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
)

func (a *Analyzer) analyze_call(call *ast.Call) tast.TypedExpr {
	callable := a.analyze_expression(call.Callable)

	fnType, ok := semantic.Resolve(callable.Type()).(semantic.FunctionType)
	if !ok {
		a.Report(
			call.Callable.NodeSpan(),
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

	for i, arg := range call.Arguments {
		typedArg := a.analyze_expression(arg)

		// This is EXACTLY the same target typing as assignment.
		typedArg, ok = coerce_to(typedArg, fnType.Params[i])
		if !ok {
			a.Report(
				arg.NodeSpan(),
				"argument %d: expected '%s', got '%s'",
				i+1,
				fnType.Params[i],
				typedArg.Type(),
			)

			return tast.ErrorExpr{
				NodeInfo: call.NodeInfo,
			}
		}

		args = append(args, typedArg)
	}

	return &tast.Call{
		NodeInfo:  call.NodeInfo,
		Callable:  callable,
		Arguments: args,
		Return:    semantic.Resolve(fnType.Return),
	}
}
