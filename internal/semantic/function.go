package semantic

import "pascc/internal/utils"

type FunctionType struct {
	Params      []Type
	ByReference []bool
	Return      Type
}

func (f FunctionType) Signature() TypeSignature {
	p := utils.Map(f.Params, func(p Type) TypeSignature {
		return p.Signature()
	})

	params := []TypeSignature{}

	params = append(params, p...)
	params = append(params, f.Return.Signature())

	return TypeSignature{
		BaseType:      "function",
		GenericParams: params,
	}
}
