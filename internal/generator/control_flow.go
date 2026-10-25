package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/utils"

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

func (g *Generator) genWhile(w *tast.While) *jen.Statement {
	return jen.For(g.generate(w.Condition)).Block(g.genStatementBlock(w.Stmts)...)
}

func (g *Generator) genCase(c *tast.Case) *jen.Statement {
	return jen.Case(g.generate(c.Value)).Block(g.genStatementBlock(c.Stmts)...)
}

func (g *Generator) genSwitch(s *tast.Switch) *jen.Statement {
	cases := utils.Map(s.Cases, func(c *tast.Case) jen.Code { return g.genCase(c) })
	_switch := jen.Switch(g.generate(s.Base)).Add(cases...)

	if len(s.Default) != 0 {
		_switch.Add(
			jen.Default().Block(g.genStatementBlock(s.Default)...),
		)
	}

	return _switch
}
