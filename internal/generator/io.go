package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/utils"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genWrite(w *tast.Write) *jen.Statement {
	f := WriteFunc()

	if !w.Newline {
		f = WriteContinuousFunc()
	}

	return f.Call(
		utils.Map(
			w.Content,
			func(e tast.TypedExpr) jen.Code { return g.generate(e) },
		)...,
	)
}

func (g *Generator) genRead(w *tast.Read) *jen.Statement {
	return ReadFunc().Call(jen.Op("&").Add(g.generate(w.Into)))
}

func (g *Generator) genClearScreen(cs *tast.ClearScreen) *jen.Statement {
	return ClearScreenFunc().Call()
}
