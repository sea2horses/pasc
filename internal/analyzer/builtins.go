package analyzer

import "pascc/internal/semantic"

func (a *Analyzer) declare_builtins() {
	integer := semantic.Integer_T{}
	real := semantic.Real_T{}
	str := semantic.String_T{}

	type builtinDecl struct {
		name   string
		params []semantic.Type
		result semantic.Type
		kind   semantic.BuiltinKind
	}
	builtins := []builtinDecl{
		{"RC", []semantic.Type{real}, real, semantic.BuiltinSqrt},
		{"Raiz", []semantic.Type{real}, real, semantic.BuiltinSqrt},
		{"Abs", []semantic.Type{real}, real, semantic.BuiltinAbs},
		{"Ln", []semantic.Type{real}, real, semantic.BuiltinLn},
		{"Exp", []semantic.Type{real}, real, semantic.BuiltinExp},
		{"Sen", []semantic.Type{real}, real, semantic.BuiltinSin},
		{"Cos", []semantic.Type{real}, real, semantic.BuiltinCos},
		{"Tan", []semantic.Type{real}, real, semantic.BuiltinTan},
		{"ASen", []semantic.Type{real}, real, semantic.BuiltinASin},
		{"ACos", []semantic.Type{real}, real, semantic.BuiltinACos},
		{"ATan", []semantic.Type{real}, real, semantic.BuiltinATan},
		{"Trunc", []semantic.Type{real}, integer, semantic.BuiltinTrunc},
		{"Redon", []semantic.Type{real}, integer, semantic.BuiltinRound},
		{"Azar", []semantic.Type{integer}, integer, semantic.BuiltinRandom},
		{"Aleatorio", []semantic.Type{integer, integer}, integer, semantic.BuiltinRandomRange},
		{"Longitud", []semantic.Type{str}, integer, semantic.BuiltinLength},
		{"Mayusculas", []semantic.Type{str}, str, semantic.BuiltinUpper},
		{"Minusculas", []semantic.Type{str}, str, semantic.BuiltinLower},
		{"Subcadena", []semantic.Type{str, integer, integer}, str, semantic.BuiltinSubstring},
		{"Concatenar", []semantic.Type{str, str}, str, semantic.BuiltinConcat},
		{"ConvertirANumero", []semantic.Type{str}, real, semantic.BuiltinToNumber},
		{"ConvertirATexto", []semantic.Type{real}, str, semantic.BuiltinToText},
		{"HoraActual", nil, integer, semantic.BuiltinCurrentTime},
		{"FechaActual", nil, integer, semantic.BuiltinCurrentDate},
	}

	for _, builtin := range builtins {
		a.scope.Declare(&semantic.Symbol{
			Name: builtin.name,
			Kind: semantic.SymbolFunction,
			Type: semantic.FunctionType{
				Params: builtin.params,
				Return: builtin.result,
			},
			Builtin: &semantic.BuiltinInfo{Kind: builtin.kind},
			Mutable: false,
		})
	}

	a.scope.Declare(&semantic.Symbol{
		Name:    "PI",
		Kind:    semantic.SymbolVariable,
		Type:    real,
		Mutable: false,
	})
}
