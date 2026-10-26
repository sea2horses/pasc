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
		if read, ok := stmt.(*tast.Read); ok && read.Declare {
			variable := read.Into.(*tast.VariableExpr)
			code = append(code, jen.Var().Id(variable.Symbol.Name).Add(g.genType(semantic.Resolve(variable.Type()))))
			code = append(code, g.generate(stmt))
			continue
		}
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
	for _, function := range mf.Functions {
		g.genFunction(function)
	}

	preload_statements := []jen.Code{jen.Defer().Id("writer").Dot("Flush").Call()}
	inner_statements := g.genStatementBlock(mf.Stmts)

	body := slices.Concat(preload_statements, inner_statements)

	f.Commentf("original pseint name: %s", mf.Name)
	f.Func().Id("main").Params().Block(body...)
}

func (g *Generator) genFunction(function *tast.Function) {
	params := make([]jen.Code, 0, len(function.Parameters))
	var body []jen.Code
	for _, parameter := range function.Parameters {
		parameterType := g.genType(parameter.Type)
		if parameter.ByReference {
			parameterType = jen.Op("*").Add(parameterType)
		}
		params = append(params, jen.Id(parameter.Name).Add(parameterType))
		body = append(body, jen.Id("_").Op("=").Id(parameter.Name))
	}
	body = append(body, g.genStatementBlock(function.Stmts)...)
	if function.HasReturn {
		body = append(body, jen.Return(g.generate(function.Return)))
	}

	declaration := g.file.Func().Id(function.Name).Params(params...)
	if function.HasReturn {
		declaration.Add(g.genType(function.Return.Type()))
	}
	declaration.Block(body...)
}
