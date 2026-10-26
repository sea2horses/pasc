package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
	"pascc/internal/utils"

	"github.com/dave/jennifer/jen"
)

var builtinMap map[semantic.BuiltinKind]func() *jen.Statement = map[semantic.BuiltinKind]func() *jen.Statement{
	semantic.BuiltinSqrt:        func() *jen.Statement { return jen.Id("Sqrt") },
	semantic.BuiltinSin:         SinFunc,
	semantic.BuiltinCos:         CosFunc,
	semantic.BuiltinTrunc:       func() *jen.Statement { return jen.Id("Trunc") },
	semantic.BuiltinAbs:         func() *jen.Statement { return jen.Id("Abs") },
	semantic.BuiltinLn:          func() *jen.Statement { return jen.Id("Ln") },
	semantic.BuiltinExp:         func() *jen.Statement { return jen.Id("Exp") },
	semantic.BuiltinTan:         func() *jen.Statement { return jen.Id("Tan") },
	semantic.BuiltinASin:        func() *jen.Statement { return jen.Id("ASin") },
	semantic.BuiltinACos:        func() *jen.Statement { return jen.Id("ACos") },
	semantic.BuiltinATan:        func() *jen.Statement { return jen.Id("ATan") },
	semantic.BuiltinRound:       func() *jen.Statement { return jen.Id("Round") },
	semantic.BuiltinRandom:      func() *jen.Statement { return jen.Id("Azar") },
	semantic.BuiltinRandomRange: func() *jen.Statement { return jen.Id("Aleatorio") },
	semantic.BuiltinLength:      func() *jen.Statement { return jen.Id("Longitud") },
	semantic.BuiltinUpper:       func() *jen.Statement { return jen.Id("Mayusculas") },
	semantic.BuiltinLower:       func() *jen.Statement { return jen.Id("Minusculas") },
	semantic.BuiltinSubstring:   func() *jen.Statement { return jen.Id("Subcadena") },
	semantic.BuiltinConcat:      func() *jen.Statement { return jen.Id("Concatenar") },
	semantic.BuiltinToNumber:    func() *jen.Statement { return jen.Id("ConvertirANumero") },
	semantic.BuiltinToText:      func() *jen.Statement { return jen.Id("ConvertirATexto") },
	semantic.BuiltinCurrentTime: func() *jen.Statement { return jen.Id("HoraActual") },
	semantic.BuiltinCurrentDate: func() *jen.Statement { return jen.Id("FechaActual") },
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
