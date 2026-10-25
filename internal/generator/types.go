package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"

	"github.com/dave/jennifer/jen"
)

var typeMap map[semantic.Type]*jen.Statement = map[semantic.Type]*jen.Statement{
	semantic.IntegerType{}: jen.Id("int64"),
	semantic.StringType{}:  jen.Id("string"),
	semantic.BooleanType{}: jen.Id("bool"),
	semantic.RealType{}:    jen.Id("float64"),
}

var conversionMap map[semantic.ConversionKind]*jen.Statement = map[semantic.ConversionKind]*jen.Statement{
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
	return code
}
