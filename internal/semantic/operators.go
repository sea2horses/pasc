package semantic

import "slices"

type Associativity uint8

const (
	LeftAssociative  Associativity = iota
	RightAssociative               // Literally only ^
)

type OperatorType uint8

const (
	ADD      OperatorType = iota // +
	SUBTRACT                     // -
	MULTIPLY                     // *
	DIVIDE                       // /
	MODULO                       // % or MOD
	POWER                        // ^

	GREATER    // >
	GREATER_EQ // >=
	LESSER     // <
	LESSER_EQ  // <=
	AND        // &&
	OR         // ||
	NOT        // !

	EQUALS     // ==
	NOT_EQUALS // !

	UNRECOGNIZED // fallback
)

func (op OperatorType) Associativity() Associativity {
	switch op {
	case POWER:
		return RightAssociative
	default:
		return LeftAssociative
	}
}

func (op OperatorType) String() string {
	switch op {
	case ADD:
		return "+"
	case SUBTRACT:
		return "-"
	case MULTIPLY:
		return "*"
	case DIVIDE:
		return "/"
	case POWER:
		return "^"
	case MODULO:
		return "%"
	case GREATER:
		return ">"
	case GREATER_EQ:
		return ">="
	case LESSER:
		return "<"
	case LESSER_EQ:
		return "<="
	case AND:
		return "&&"
	case OR:
		return "||"
	case NOT:
		return "!"
	case EQUALS:
		return "=="
	case NOT_EQUALS:
		return "!="
	default:
		return "[unknown]"
	}
}

var unaries []OperatorType = []OperatorType{SUBTRACT, NOT}
var binaries []OperatorType = []OperatorType{ADD, SUBTRACT, MULTIPLY, DIVIDE, MODULO, POWER, GREATER, GREATER_EQ, LESSER, LESSER_EQ, AND, OR, EQUALS, NOT_EQUALS}

func IsUnary(op OperatorType) bool {
	return slices.Contains(unaries, op)
}

func IsBinary(op OperatorType) bool {
	return slices.Contains(binaries, op)
}

func BinaryOperatorPrecedence(op OperatorType) int8 {
	if !IsBinary(op) {
		return -1
	}

	switch op {
	case POWER:
		return 15
	case MULTIPLY, DIVIDE, MODULO:
		return 14
	case ADD, SUBTRACT:
		return 13
	case GREATER, GREATER_EQ, LESSER, LESSER_EQ:
		return 12
	case EQUALS, NOT_EQUALS:
		return 11
	case AND:
		return 10
	case OR:
		return 9
	}

	return -1
}
