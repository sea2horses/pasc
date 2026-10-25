package generator

import (
	tast "pascc/internal/ast_typed"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genIf(i *tast.If) *jen.Statement {
	_if := jen.If(g.generate(i.Condition)).Block(g.genStatementBlock(i.Stmts)...)
	if i.Else != nil {
		_if.Else().Block(
			g.genStatementBlock(i.Else.Stmts)...,
		)
	}
	return _if
}
