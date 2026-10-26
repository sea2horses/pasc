package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"

	"github.com/dave/jennifer/jen"
)

var typeMap map[semantic.Type]*jen.Statement = map[semantic.Type]*jen.Statement{
	semantic.Integer_T{}: jen.Id("int32"),
	semantic.String_T{}:  jen.Id("string"),
	semantic.Bool_T{}:    jen.Id("bool"),
	semantic.Real_T{}:    jen.Id("float64"),
}

var conversionMap map[semantic.ConversionKind]func() *jen.Statement = map[semantic.ConversionKind]func() *jen.Statement{
	semantic.ConversionIntegerToReal:      IntToRealFunc,
	semantic.ConversionRealToIntegerExact: ToIntegerExFunc,
}

func (g *Generator) genType(t semantic.Type) *jen.Statement {
	code, ok := typeMap[t]
	if !ok {
		panic("Type " + t.Signature().String() + " does not have a mapped representation")
	}
	return code
}

func (g *Generator) genCast(c *tast.Cast) *jen.Statement {
	code, ok := conversionMap[*c.ConversionKind]
	if !ok {
		panic("ConversionKind does not have a mapped representation")
	}
	return code().Call(g.generate(c.Expr))
}
