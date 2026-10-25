package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/utils"

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
