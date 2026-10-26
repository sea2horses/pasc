package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
	"slices"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genStatementBlock(stmts []tast.Stmt) []jen.Code {
	var code []jen.Code
	for _, stmt := range stmts {
		declaration, ok := stmt.(*tast.Declaration)
		if !ok {
			code = append(code, g.generate(stmt))
			continue
		}

		code = append(code, g.genDeclaration(declaration))
		for _, symbol := range declaration.Symbols {
			// Dimensioned names are declared by NewTensor. Mark only emitted
			// scalar declarations used so Go accepts unused PSeInt variables.
			if !semantic.IsArray(symbol.Type) {
				code = append(code, jen.Id("_").Op("=").Id(symbol.Name))
			}
		}
	}
	return code
}

func (g *Generator) genMainFunction(mf *tast.MainFunction) {
	f := g.file

	preload_statements := []jen.Code{jen.Defer().Id("writer").Dot("Flush").Call()}
	inner_statements := g.genStatementBlock(mf.Stmts)

	body := slices.Concat(preload_statements, inner_statements)

	f.Commentf("original pseint name: %s", mf.Name)
	f.Func().Id("main").Params().Block(body...)
}
