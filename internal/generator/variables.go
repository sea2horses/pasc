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
	return jen.Var().List(
		utils.Map(d.Symbols, func(s *semantic.Symbol) jen.Code {
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
