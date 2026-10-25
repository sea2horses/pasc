package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"

	"github.com/dave/jennifer/jen"
)

var timeMap map[semantic.TimeUnit]*jen.Statement = map[semantic.TimeUnit]*jen.Statement{
	semantic.MILISEGUNDOS: jen.Qual("time", "Millisecond"),
	semantic.SEGUNDOS:     jen.Qual("time", "Second"),
}

func (g *Generator) genTimeUnit(t semantic.TimeUnit) *jen.Statement {
	code, ok := timeMap[t]
	if !ok {
		panic("Timeunit does not have a mapped representation")
	}
	return code
}

func (g *Generator) genTimeout(to *tast.TimeOut) *jen.Statement {
	return TimeoutFunc.Call(
		g.generate(to.Amount).Op("*").Add(g.genTimeUnit(to.Unit)),
	)
}
