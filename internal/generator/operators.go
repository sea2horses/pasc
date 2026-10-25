package generator

import (
	"pascc/internal/semantic"

	"github.com/dave/jennifer/jen"
)

type GenerationStrategy func(a, b *jen.Statement) jen.Code

func operatorStrategy(op semantic.OperatorType) GenerationStrategy {
	switch op {
	/* Power is CURRENTLY the only operator that needs a different strategy */
	case semantic.POWER:
		return func(a, b *jen.Statement) jen.Code {
			return PowerFunc.Call(a, b)
		}

	default:
		/* By default, all operators are just expr 'op' expr */
		return func(a, b *jen.Statement) jen.Code {
			return a.Op(op.String()).Add(b)
		}
	}
}
