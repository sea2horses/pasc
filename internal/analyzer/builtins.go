package analyzer

import "pascc/internal/semantic"

func (a *Analyzer) declare_builtins() {

	builtins := []*semantic.Symbol{
		&semantic.Symbol{
			Name: "Sen",
			Kind: semantic.SymbolFunction,
			Type: semantic.FunctionType{
				Params: []semantic.Type{semantic.Real_T{}},
				Return: semantic.Real_T{},
			},
			Builtin: &semantic.BuiltinInfo{
				Kind: semantic.BuiltinSin,
			},
			Mutable: false,
		},
		&semantic.Symbol{
			Name: "Cos",
			Kind: semantic.SymbolFunction,
			Type: semantic.FunctionType{
				Params: []semantic.Type{semantic.Real_T{}},
				Return: semantic.Real_T{},
			},
			Builtin: &semantic.BuiltinInfo{
				Kind: semantic.BuiltinCos,
			},
			Mutable: false,
		},
		&semantic.Symbol{
			Name: "Trunc",
			Kind: semantic.SymbolFunction,
			Type: semantic.FunctionType{
				Params: []semantic.Type{semantic.Real_T{}},
				Return: semantic.Real_T{},
			},
			Builtin: &semantic.BuiltinInfo{
				Kind: semantic.BuiltinTrunc,
			},
			Mutable: false,
		},
	}

	for _, builtin := range builtins {
		a.scope.Declare(builtin)
	}
}
