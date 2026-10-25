package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
	"pascc/internal/utils"

	"github.com/dave/jennifer/jen"
)

var builtinMap map[semantic.BuiltinKind]func() *jen.Statement = map[semantic.BuiltinKind]func() *jen.Statement{
	semantic.BuiltinSin:   SinFunc,
	semantic.BuiltinCos:   CosFunc,
	semantic.BuiltinTrunc: TruncFunc,
}

func (g *Generator) genBuiltinCall(bi *semantic.BuiltinInfo, args []tast.TypedExpr) *jen.Statement {
	f, ok := builtinMap[bi.Kind]
	if !ok {
		panic("BuiltinKind does not have a mapped representation")
	}
	return f().Call(
		utils.Map(args, func(expr tast.TypedExpr) jen.Code { return g.generate(expr) })...,
	)
}

func (g *Generator) genCall(c *tast.Call) *jen.Statement {
	if variable, ok := c.Callable.(*tast.VariableExpr); ok && variable.Symbol.Builtin != nil {
		return g.genBuiltinCall(variable.Symbol.Builtin, c.Arguments)
	}
	return g.generate(c.Callable).Call(
		utils.Map(c.Arguments, func(expr tast.TypedExpr) jen.Code { return g.generate(expr) })...,
	)
}
