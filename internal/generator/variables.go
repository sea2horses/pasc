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
	var op *jen.Statement

	if a.Declarative {
		op = jen.Op(":=")
	} else {
		op = jen.Op("=")
	}

	return g.generate(a.Target).Add(op, g.generate(a.Value))
}
