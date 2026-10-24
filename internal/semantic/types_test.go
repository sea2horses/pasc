package semantic_test

import (
	"pascc/internal/semantic"
	"testing"
)

func TestEqualities(t *testing.T) {
	/*  */
	intType := semantic.IntegerType{}

	if !semantic.IsInteger(intType) {
		t.Errorf("expected intType to be an integer type")
	}

	realType := semantic.RealType{}

	if !semantic.IsReal(realType) {
		t.Errorf("expected realType to be a real type")
	}

	stringType := semantic.StringType{}

	if !semantic.IsString(stringType) {
		t.Errorf("expected stringType to be a string type")
	}

	boolType := semantic.BooleanType{}

	if !semantic.IsBoolean(boolType) {
		t.Errorf("expected boolType to be a bool type")
	}
}

func TestIsVoid(t *testing.T) {
	voidType := semantic.VoidType{}

	if !semantic.IsVoid(voidType) {
		t.Errorf("expected voidType to be a void type")
	}
}

func TestInvalid(t *testing.T) {
	invalidType := semantic.InvalidType{}

	if !semantic.IsInvalid(invalidType) {
		t.Errorf("expected invalidType to be an invalid type")
	}
}

func TestArrays(t *testing.T) {
	arrayType := semantic.ArrayType{Elem: semantic.IntegerType{}}

	if !semantic.IsArray(arrayType) {
		t.Errorf("expected arrayType to be an array type")
	}
}
