package generator

import (
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
	"pascc/internal/utils"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genDimension(d *tast.Dimension) *jen.Statement {
	_underlying := d.Symbol.Type.(semantic.ArrayType)
	_type := semantic.Resolve(_underlying.Elem)

	return jen.Id(d.Symbol.Name).Op(":=").Add(MakeTensorFunc).Types(g.genType(_type)).Call(
		utils.Map(d.Dimensions, func(i tast.TypedExpr) jen.Code { return jen.Int().Parens(g.generate(i)) })...,
	)
}

func (g *Generator) genIndex(i *tast.Index) *jen.Statement {
	args := []jen.Code{g.generate(i.Target)}
	args = append(args, utils.Map(i.Indexes, func(i tast.TypedExpr) jen.Code { return jen.Int().Parens(g.generate(i)) })...)
	return jen.Op("*").Add(IndexTensor).Call(args...)
}
