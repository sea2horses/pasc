package semantic

import (
	"fmt"
	"pascc/internal/utils"
	"strings"
)

type TypeSignature struct {
	BaseType      string
	GenericParams []TypeSignature
	Metadata      string
}

func (t TypeSignature) String() string {
	var s strings.Builder

	s.WriteString(t.BaseType)
	if len(t.GenericParams) > 0 || t.Metadata != "" {
		params := strings.Join(utils.Map(t.GenericParams, func(gp TypeSignature) string {
			return gp.String()
		}), ", ")

		final := strings.Join([]string{params, t.Metadata}, ";")

		fmt.Fprintf(&s, "<%s>", final)
	}

	return s.String()
}

func (t TypeSignature) Equals(other TypeSignature) bool {
	if t.BaseType != other.BaseType {
		return false
	}
	if len(t.GenericParams) != len(other.GenericParams) {
		return false
	}
	for i := range t.GenericParams {
		if !t.GenericParams[i].Equals(other.GenericParams[i]) {
			return false
		}
	}
	return true
}

type Type interface{ Signature() TypeSignature }

/* Based on type signatures */
func EqualTypes(a, b Type) bool {
	return a.Signature().Equals(b.Signature())
}

/* Checks if two types are equal at the top level (array vs array) */
func EqualTopLevelType(a, b Type) bool {
	return a.Signature().BaseType == b.Signature().BaseType
}

func IsInvalid(t Type) bool {
	prim, ok := t.(PrimitiveType)
	return ok && prim.Kind() == INVALID
}

func IsVoid(t Type) bool {
	return EqualTypes(t, Void_T{})
}

func IsNumeric(t Type) bool {
	prim, ok := t.(PrimitiveType)
	if !ok {
		return false
	}

	return prim.Kind() == UNTYPED_NUMBER || prim.Kind() == INTEGER || prim.Kind() == REAL
}

func IsInteger(t Type) bool {
	return EqualTypes(t, Integer_T{})
}

func IsReal(t Type) bool {
	return EqualTypes(t, Real_T{})
}

func IsString(t Type) bool {
	return EqualTypes(t, String_T{})
}

func IsBoolean(t Type) bool {
	return EqualTypes(t, Bool_T{})
}

func IsGeneric(t Type) bool {
	return len(t.Signature().GenericParams) > 0
}

func IsArray(t Type) bool {
	return EqualTopLevelType(t, ArrayType{Elem: Void_T{}})
}
