package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/i18n"
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
		panic(i18n.T(i18n.CodegenUnmapped, "TimeUnit", t))
	}
	return code
}

func (g *Generator) genTimeout(to *tast.TimeOut) *jen.Statement {
	// Convert both operands to float64 before multiplication so fractional
	// delays (for example, Esperar 0.5 Segundos) remain valid Go expressions.
	duration := jen.Qual("time", "Duration").Call(
		jen.Id("float64").Call(g.generate(to.Amount)).Op("*").Add(
			jen.Id("float64").Call(g.genTimeUnit(to.Unit)),
		),
	)
	return TimeoutFunc().Call(duration)
}
