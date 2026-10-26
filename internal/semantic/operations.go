package semantic

type ConversionKind uint8

const (
	ConversionNone ConversionKind = iota
	ConversionIntegerToReal
	ConversionRealToIntegerExact
)

type ImplicitCast struct {
	From           Type
	To             Type
	ConversionKind ConversionKind
}

type BinaryOperationRule struct {
	Operator        OperatorType
	Left            Type
	Right           Type
	Result          Type
	LeftConversion  ConversionKind
	RightConversion ConversionKind
}

type UnaryOperationRule struct {
	Operator   OperatorType
	Operand    Type
	Result     Type
	Conversion ConversionKind
}

var allowedCasts = []ImplicitCast{
	{From: Integer_T{}, To: Real_T{}, ConversionKind: ConversionIntegerToReal},
	{From: Real_T{}, To: Integer_T{}, ConversionKind: ConversionRealToIntegerExact},
}

var binaryOperationTable = []BinaryOperationRule{
	// Addition and string concatenation.
	{ADD, Integer_T{}, Integer_T{}, Integer_T{}, ConversionNone, ConversionNone},
	{ADD, Integer_T{}, Real_T{}, Real_T{}, ConversionIntegerToReal, ConversionNone},
	{ADD, Real_T{}, Integer_T{}, Real_T{}, ConversionNone, ConversionIntegerToReal},
	{ADD, Real_T{}, Real_T{}, Real_T{}, ConversionNone, ConversionNone},
	{ADD, String_T{}, String_T{}, String_T{}, ConversionNone, ConversionNone},

	// Arithmetic.
	{SUBTRACT, Integer_T{}, Integer_T{}, Integer_T{}, ConversionNone, ConversionNone},
	{SUBTRACT, Integer_T{}, Real_T{}, Real_T{}, ConversionIntegerToReal, ConversionNone},
	{SUBTRACT, Real_T{}, Integer_T{}, Real_T{}, ConversionNone, ConversionIntegerToReal},
	{SUBTRACT, Real_T{}, Real_T{}, Real_T{}, ConversionNone, ConversionNone},

	{MULTIPLY, Integer_T{}, Integer_T{}, Integer_T{}, ConversionNone, ConversionNone},
	{MULTIPLY, Integer_T{}, Real_T{}, Real_T{}, ConversionIntegerToReal, ConversionNone},
	{MULTIPLY, Real_T{}, Integer_T{}, Real_T{}, ConversionNone, ConversionIntegerToReal},
	{MULTIPLY, Real_T{}, Real_T{}, Real_T{}, ConversionNone, ConversionNone},

	// Division and exponentiation always produce a real value.
	{DIVIDE, Integer_T{}, Integer_T{}, Real_T{}, ConversionIntegerToReal, ConversionIntegerToReal},
	{DIVIDE, Integer_T{}, Real_T{}, Real_T{}, ConversionIntegerToReal, ConversionNone},
	{DIVIDE, Real_T{}, Integer_T{}, Real_T{}, ConversionNone, ConversionIntegerToReal},
	{DIVIDE, Real_T{}, Real_T{}, Real_T{}, ConversionNone, ConversionNone},

	{POWER, Integer_T{}, Integer_T{}, Real_T{}, ConversionIntegerToReal, ConversionIntegerToReal},
	{POWER, Integer_T{}, Real_T{}, Real_T{}, ConversionIntegerToReal, ConversionNone},
	{POWER, Real_T{}, Integer_T{}, Real_T{}, ConversionNone, ConversionIntegerToReal},
	{POWER, Real_T{}, Real_T{}, Real_T{}, ConversionNone, ConversionNone},

	// MOD is strict and integer-only.
	{MODULO, Integer_T{}, Integer_T{}, Integer_T{}, ConversionNone, ConversionNone},
	{MODULO, Integer_T{}, Real_T{}, Integer_T{}, ConversionNone, ConversionRealToIntegerExact},
	{MODULO, Real_T{}, Integer_T{}, Real_T{}, ConversionRealToIntegerExact, ConversionNone},
	{MODULO, Real_T{}, Real_T{}, Integer_T{}, ConversionRealToIntegerExact, ConversionRealToIntegerExact},

	// Numeric ordering.
	{GREATER, Integer_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{GREATER, Integer_T{}, Real_T{}, Bool_T{}, ConversionIntegerToReal, ConversionNone},
	{GREATER, Real_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionIntegerToReal},
	{GREATER, Real_T{}, Real_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{GREATER_EQ, Integer_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{GREATER_EQ, Integer_T{}, Real_T{}, Bool_T{}, ConversionIntegerToReal, ConversionNone},
	{GREATER_EQ, Real_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionIntegerToReal},
	{GREATER_EQ, Real_T{}, Real_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{LESSER, Integer_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{LESSER, Integer_T{}, Real_T{}, Bool_T{}, ConversionIntegerToReal, ConversionNone},
	{LESSER, Real_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionIntegerToReal},
	{LESSER, Real_T{}, Real_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{LESSER_EQ, Integer_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{LESSER_EQ, Integer_T{}, Real_T{}, Bool_T{}, ConversionIntegerToReal, ConversionNone},
	{LESSER_EQ, Real_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionIntegerToReal},
	{LESSER_EQ, Real_T{}, Real_T{}, Bool_T{}, ConversionNone, ConversionNone},

	// Equality is intentionally unavailable for strings.
	{EQUALS, Integer_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{EQUALS, Integer_T{}, Real_T{}, Bool_T{}, ConversionIntegerToReal, ConversionNone},
	{EQUALS, Real_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionIntegerToReal},
	{EQUALS, Real_T{}, Real_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{EQUALS, Bool_T{}, Bool_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{NOT_EQUALS, Integer_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{NOT_EQUALS, Integer_T{}, Real_T{}, Bool_T{}, ConversionIntegerToReal, ConversionNone},
	{NOT_EQUALS, Real_T{}, Integer_T{}, Bool_T{}, ConversionNone, ConversionIntegerToReal},
	{NOT_EQUALS, Real_T{}, Real_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{NOT_EQUALS, Bool_T{}, Bool_T{}, Bool_T{}, ConversionNone, ConversionNone},

	// Logical
	{AND, Bool_T{}, Bool_T{}, Bool_T{}, ConversionNone, ConversionNone},
	{OR, Bool_T{}, Bool_T{}, Bool_T{}, ConversionNone, ConversionNone},
}

var UnaryOperationTable = []UnaryOperationRule{
	{SUBTRACT, Integer_T{}, Integer_T{}, ConversionNone},
	{SUBTRACT, Real_T{}, Real_T{}, ConversionNone},
	{NOT, Bool_T{}, Bool_T{}, ConversionNone},
}

func CanCast(from Type, to Type) *ConversionKind {
	for _, cast := range allowedCasts {
		if EqualTypes(cast.From, from) && EqualTypes(cast.To, to) {
			return &cast.ConversionKind
		}
	}
	return nil
}

// func SearchConversion(with Type, c ConversionKind) *ImplicitCast {
// 	for _, cast := range allowedCasts {
// 		if EqualTypes(cast.From, with) && cast.ConversionKind == c {
// 			return &cast
// 		}
// 	}
// 	return nil
// }

func ResolveBinaryOperation(op OperatorType, left, right Type) (BinaryOperationRule, bool) {
	for _, rule := range binaryOperationTable {
		if rule.Operator == op && EqualTypes(rule.Left, left) && EqualTypes(rule.Right, right) {
			return rule, true
		}
	}

	return BinaryOperationRule{}, false
}

func ResolveUnaryOperation(op OperatorType, operand Type) (UnaryOperationRule, bool) {
	for _, rule := range UnaryOperationTable {
		if rule.Operator == op && EqualTypes(rule.Operand, operand) {
			return rule, true
		}
	}

	return UnaryOperationRule{}, false
}
