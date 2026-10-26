package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"

	"github.com/dave/jennifer/jen"
)

func IntegerType() *jen.Statement { return jen.Id("int32") }
func StringType() *jen.Statement  { return jen.Id("string") }
func BoolType() *jen.Statement    { return jen.Id("bool") }
func RealType() *jen.Statement    { return jen.Id("float64") }

var typeMap map[semantic.Type]func() *jen.Statement = map[semantic.Type]func() *jen.Statement{
	semantic.Integer_T{}: IntegerType,
	semantic.String_T{}:  StringType,
	semantic.Bool_T{}:    BoolType,
	semantic.Real_T{}:    RealType,
}

var conversionMap map[semantic.ConversionKind]func() *jen.Statement = map[semantic.ConversionKind]func() *jen.Statement{
	semantic.ConversionIntegerToReal:      IntToRealFunc,
	semantic.ConversionRealToIntegerExact: ToIntegerExFunc,
}

func (g *Generator) genType(t semantic.Type) *jen.Statement {
	t = semantic.Resolve(t)
	code, ok := typeMap[t]
	if !ok {
		panic("Type " + t.Signature().String() + " does not have a mapped representation")
	}
	return code()
}

func (g *Generator) genCast(c *tast.Cast) *jen.Statement {
	code, ok := conversionMap[*c.ConversionKind]
	if !ok {
		panic("ConversionKind does not have a mapped representation")
	}
	return code().Call(g.generate(c.Expr))
}
