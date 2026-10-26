package generator

import (
	"fmt"
	"strconv"
	"strings"

	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"

	"github.com/dave/jennifer/jen"
)

func (g *Generator) genNumberLiteral(nl *tast.NumberLiteral) *jen.Statement {
	switch semantic.Resolve(nl.Type()).(type) {
	case semantic.Integer_T:
		if strings.Trim(nl.Frac, "0") != "" {
			panic("non-integral number literal resolved as integer")
		}
		return genInt(nl)

	case semantic.Real_T:
		return genRealLiteral(nl)

	case semantic.UntypedNumber_T:
		// Untyped literals should normally be settled by the analyzer. Keep
		// code generation deterministic if an unconstrained one reaches here.
		if nl.Frac == "" {
			return genInt(nl)
		}
		return genRealLiteral(nl)

	default:
		panic(fmt.Sprintf("number literal has non-numeric type %s", nl.Type()))
	}
}

func genInt(nl *tast.NumberLiteral) *jen.Statement {
	text := nl.Int
	/* raw printing */
	return jen.Id(text)
}

func genRealLiteral(nl *tast.NumberLiteral) *jen.Statement {
	text := nl.Int
	if nl.Frac != "" {
		text += "." + nl.Frac
	}
	num, _ := strconv.ParseFloat(text, 64)
	return jen.Lit(num)
}

func (g *Generator) genStringLiteral(sl *tast.StringLiteral) *jen.Statement {
	return jen.Lit(sl.Value)
}

func (g *Generator) genBoolLiteral(bl *tast.BooleanLiteral) *jen.Statement {
	return jen.Lit(bl.Value)
}
