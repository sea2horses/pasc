package generator

import (
	tast "pascc/internal/ast_typed"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genVariable(v *tast.VariableExpr) *jen.Statement {
	return jen.Id(v.Symbol.Name)
}

func (g *Generator) genUnExpr(ue *tast.UnaryOperation) *jen.Statement {
	return jen.Parens(jen.Op(ue.Op.Op.String()).Add(g.generate(ue.Expr)))
}

func (g *Generator) genBinExpr(be *tast.BinaryOperation) *jen.Statement {
	/* Get the generation strategy for the operator */
	strategy := operatorStrategy(be.Op.Op)
	return jen.Parens(strategy(g.generate(be.LHS), g.generate(be.RHS)))
}
