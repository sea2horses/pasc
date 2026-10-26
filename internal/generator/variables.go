package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
	"pascc/internal/utils"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genVariable(v *tast.VariableExpr) *jen.Statement {
	return jen.Id(v.Symbol.Name)
}

func (g *Generator) genDeclaration(d *tast.Declaration) *jen.Statement {
	var symbols []*semantic.Symbol
	for _, symbol := range d.Symbols {
		// A scalar definition can later be refined by Dimensionar into an
		// array. The array allocation declares that variable instead.
		if !semantic.IsArray(symbol.Type) {
			symbols = append(symbols, symbol)
		}
	}
	if len(symbols) == 0 {
		return jen.Empty()
	}
	return jen.Var().List(
		utils.Map(symbols, func(s *semantic.Symbol) jen.Code {
			return jen.Id(s.Name)
		})...,
	).Add(g.genType(d.Type))
}

func (g *Generator) genAssignment(a *tast.Assignment) *jen.Statement {
	code := g.generate(a.Target)
	var op *jen.Statement
	value := g.generate(a.Value)

	if a.Declarative {
		op = jen.Op(":=")
		/* explicit typign for numeric types */
		if semantic.IsNumeric(a.Value.Type()) {
			value = g.genType(a.Value.Type()).Call(value)
		}
	} else {
		op = jen.Op("=")
	}

	return code.Add(op, value)
}
