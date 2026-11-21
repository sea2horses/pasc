package analyzer

import "pascc/internal/semantic"

type builtinDecl struct {
	name   string
	params []semantic.Type
	result semantic.Type
	kind   semantic.BuiltinKind
}

var (
	tInteger = semantic.Integer_T{}
	tReal    = semantic.Real_T{}
	tString  = semantic.String_T{}
)

var builtins = []builtinDecl{
	{"RC", []semantic.Type{tReal}, tReal, semantic.BuiltinSqrt},
	{"Raiz", []semantic.Type{tReal}, tReal, semantic.BuiltinSqrt},
	{"Abs", []semantic.Type{tReal}, tReal, semantic.BuiltinAbs},
	{"Ln", []semantic.Type{tReal}, tReal, semantic.BuiltinLn},
	{"Exp", []semantic.Type{tReal}, tReal, semantic.BuiltinExp},
	{"Sen", []semantic.Type{tReal}, tReal, semantic.BuiltinSin},
	{"Cos", []semantic.Type{tReal}, tReal, semantic.BuiltinCos},
	{"Tan", []semantic.Type{tReal}, tReal, semantic.BuiltinTan},
	{"ASen", []semantic.Type{tReal}, tReal, semantic.BuiltinASin},
	{"ACos", []semantic.Type{tReal}, tReal, semantic.BuiltinACos},
	{"ATan", []semantic.Type{tReal}, tReal, semantic.BuiltinATan},
	{"Trunc", []semantic.Type{tReal}, tInteger, semantic.BuiltinTrunc},
	{"Redon", []semantic.Type{tReal}, tInteger, semantic.BuiltinRound},
	{"Redondear", []semantic.Type{tReal}, tInteger, semantic.BuiltinRound},
	{"Azar", []semantic.Type{tInteger}, tInteger, semantic.BuiltinRandom},
	{"Aleatorio", []semantic.Type{tInteger, tInteger}, tInteger, semantic.BuiltinRandomRange},
	{"Longitud", []semantic.Type{tString}, tInteger, semantic.BuiltinLength},
	{"Mayusculas", []semantic.Type{tString}, tString, semantic.BuiltinUpper},
	{"Minusculas", []semantic.Type{tString}, tString, semantic.BuiltinLower},
	{"Subcadena", []semantic.Type{tString, tInteger, tInteger}, tString, semantic.BuiltinSubstring},
	{"Concatenar", []semantic.Type{tString, tString}, tString, semantic.BuiltinConcat},
	{"ConvertirANumero", []semantic.Type{tString}, tReal, semantic.BuiltinToNumber},
	{"ConvertirATexto", []semantic.Type{tReal}, tString, semantic.BuiltinToText},
	{"HoraActual", nil, tInteger, semantic.BuiltinCurrentTime},
	{"FechaActual", nil, tInteger, semantic.BuiltinCurrentDate},
}

// builtinConstants are predeclared read-only variables.
var builtinConstants = []string{"PI"}

// IsBuiltin reports whether name (in any casing) is a predeclared function or
// constant, such as Raiz or PI.
func IsBuiltin(name string) bool {
	name = semantic.NormalizeName(name)
	for _, builtin := range builtins {
		if semantic.NormalizeName(builtin.name) == name {
			return true
		}
	}
	for _, constant := range builtinConstants {
		if semantic.NormalizeName(constant) == name {
			return true
		}
	}
	return false
}

func (a *Analyzer) declare_builtins() {
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

	for _, constant := range builtinConstants {
		a.scope.Declare(&semantic.Symbol{
			Name:    constant,
			Kind:    semantic.SymbolVariable,
			Type:    tReal,
			Mutable: false,
		})
	}
}
