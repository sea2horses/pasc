package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/utils"
	"slices"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genStatementBlock(stmts []tast.Stmt) []jen.Code {
	return utils.Map(
		stmts,
		func(stmt tast.Stmt) jen.Code {
			return g.generate(stmt)
		},
	)
}

func (g *Generator) genMainFunction(mf *tast.MainFunction) {
	f := g.file

	preload_statements := []jen.Code{jen.Defer().Id("writer").Dot("Flush").Call()}
	inner_statements := g.genStatementBlock(mf.Stmts)

	body := slices.Concat(preload_statements, inner_statements)

	f.Commentf("original pseint name: %s", mf.Name)
	f.Func().Id("main").Params().Block(body...)
}
