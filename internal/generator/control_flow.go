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
	if len(s.Default) != 0 {
		cases = append(cases, jen.Default().Block(g.genStatementBlock(s.Default)...))
	}

	_switch := jen.Switch(g.generate(s.Base)).Block(cases...)

	return _switch
}

func (g *Generator) genFor(f *tast.For) *jen.Statement {
	var _range *jen.Statement
	if f.Step == nil {
		_range = IterSteplessFunc.Call(g.generate(f.Start), g.generate(f.Until))
	} else {
		_range = IterFunc.Call(g.generate(f.Start), g.generate(f.Until), g.generate(f.Step))
	}

	_for := jen.For(
		g.generate(f.Var).Op(":=").Add(_range),
	).Block(
		g.genStatementBlock(f.Stmts)...,
	)

	return jen.Block(_for)
}
