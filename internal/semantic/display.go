package semantic

import "pascc/internal/i18n"

// DisplayType names a type the way a PSeInt programmer would write it, for
// use in diagnostics. String()/Signature() stay internal and stable.
func DisplayType(t Type) string {
	switch t := Resolve(t).(type) {
	case nil:
		return i18n.T(i18n.TypeUnknown)
	case Integer_T:
		return "Entero"
	case Real_T:
		return "Real"
	case UntypedNumber_T:
		return "Numero"
	case String_T:
		return "Cadena"
	case Bool_T:
		return "Logico"
	case Void_T:
		return i18n.T(i18n.TypeVoid)
	case ArrayType:
		if t.Rank <= 1 {
			return i18n.T(i18n.TypeArray, DisplayType(t.Elem))
		}
		return i18n.T(i18n.TypeArrayND, DisplayType(t.Elem), t.Rank)
	case FunctionType:
		return i18n.T(i18n.TypeFunction)
	case *InferType:
		return i18n.T(i18n.TypeUnknown)
	}
	return t.Signature().String()
}

// Display returns the operator as written in PSeInt source.
func (op OperatorType) Display() string {
	switch op {
	case AND:
		return "Y"
	case OR:
		return "O"
	case NOT:
		return "NO"
	case EQUALS:
		return "="
	case NOT_EQUALS:
		return "<>"
	case MODULO:
		return "MOD"
	}
	return op.String()
}
