package generator

import (
	tast "pascc/internal/ast_typed"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genVariable(v *tast.VariableExpr) *jen.Statement {
	return jen.Id(v.Symbol.Name)
}
