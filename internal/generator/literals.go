package generator

import (
	tast "pascc/internal/ast_typed"
	"strconv"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genNumberLiteral(nl *tast.NumberLiteral) *jen.Statement {
	if nl.Frac != "" {
		num, _ := strconv.ParseFloat(nl.Int+"."+nl.Frac, 64)
		return jen.Lit(num)
	} else {
		/* Until I don't do constexpr evaluation, this has to be float too */
		num, _ := strconv.ParseFloat(nl.Int, 64)
		return jen.Lit(num)
	}
}

func (g *Generator) genStringLiteral(sl *tast.StringLiteral) *jen.Statement {
	return jen.Lit(sl.Value)
}

func (g *Generator) genBoolLiteral(bl *tast.BooleanLiteral) *jen.Statement {
	return jen.Lit(bl.Value)
}
