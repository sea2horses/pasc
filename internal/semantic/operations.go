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

var AllowedCasts = []ImplicitCast{
	{From: IntegerType{}, To: RealType{}, ConversionKind: ConversionIntegerToReal},
	{From: RealType{}, To: IntegerType{}, ConversionKind: ConversionRealToIntegerExact},
}

var BinaryOperationTable = []BinaryOperationRule{
	// Addition and string concatenation.
	{ADD, IntegerType{}, IntegerType{}, IntegerType{}, ConversionNone, ConversionNone},
	{ADD, IntegerType{}, RealType{}, RealType{}, ConversionIntegerToReal, ConversionNone},
	{ADD, RealType{}, IntegerType{}, RealType{}, ConversionNone, ConversionIntegerToReal},
	{ADD, RealType{}, RealType{}, RealType{}, ConversionNone, ConversionNone},
	{ADD, StringType{}, StringType{}, StringType{}, ConversionNone, ConversionNone},

	// Arithmetic.
	{SUBTRACT, IntegerType{}, IntegerType{}, IntegerType{}, ConversionNone, ConversionNone},
	{SUBTRACT, IntegerType{}, RealType{}, RealType{}, ConversionIntegerToReal, ConversionNone},
	{SUBTRACT, RealType{}, IntegerType{}, RealType{}, ConversionNone, ConversionIntegerToReal},
	{SUBTRACT, RealType{}, RealType{}, RealType{}, ConversionNone, ConversionNone},

	{MULTIPLY, IntegerType{}, IntegerType{}, IntegerType{}, ConversionNone, ConversionNone},
	{MULTIPLY, IntegerType{}, RealType{}, RealType{}, ConversionIntegerToReal, ConversionNone},
	{MULTIPLY, RealType{}, IntegerType{}, RealType{}, ConversionNone, ConversionIntegerToReal},
	{MULTIPLY, RealType{}, RealType{}, RealType{}, ConversionNone, ConversionNone},

	// Division and exponentiation always produce a real value.
	{DIVIDE, IntegerType{}, IntegerType{}, RealType{}, ConversionIntegerToReal, ConversionIntegerToReal},
	{DIVIDE, IntegerType{}, RealType{}, RealType{}, ConversionIntegerToReal, ConversionNone},
	{DIVIDE, RealType{}, IntegerType{}, RealType{}, ConversionNone, ConversionIntegerToReal},
	{DIVIDE, RealType{}, RealType{}, RealType{}, ConversionNone, ConversionNone},

	// Power will be shelved since it needs a function
	// {POWER, IntegerType{}, IntegerType{}, RealType{}, ConversionIntegerToReal, ConversionIntegerToReal},
	// {POWER, IntegerType{}, RealType{}, RealType{}, ConversionIntegerToReal, ConversionNone},
	// {POWER, RealType{}, IntegerType{}, RealType{}, ConversionNone, ConversionIntegerToReal},
	// {POWER, RealType{}, RealType{}, RealType{}, ConversionNone, ConversionNone},

	// MOD is strict and integer-only.
	{MODULO, IntegerType{}, IntegerType{}, IntegerType{}, ConversionNone, ConversionNone},
	{MODULO, IntegerType{}, RealType{}, IntegerType{}, ConversionNone, ConversionRealToIntegerExact},
	{MODULO, RealType{}, IntegerType{}, RealType{}, ConversionRealToIntegerExact, ConversionNone},
	{MODULO, RealType{}, RealType{}, IntegerType{}, ConversionRealToIntegerExact, ConversionRealToIntegerExact},

	// Numeric ordering.
	{GREATER, IntegerType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionNone},
	{GREATER, IntegerType{}, RealType{}, BooleanType{}, ConversionIntegerToReal, ConversionNone},
	{GREATER, RealType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionIntegerToReal},
	{GREATER, RealType{}, RealType{}, BooleanType{}, ConversionNone, ConversionNone},
	{GREATER_EQ, IntegerType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionNone},
	{GREATER_EQ, IntegerType{}, RealType{}, BooleanType{}, ConversionIntegerToReal, ConversionNone},
	{GREATER_EQ, RealType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionIntegerToReal},
	{GREATER_EQ, RealType{}, RealType{}, BooleanType{}, ConversionNone, ConversionNone},
	{LESSER, IntegerType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionNone},
	{LESSER, IntegerType{}, RealType{}, BooleanType{}, ConversionIntegerToReal, ConversionNone},
	{LESSER, RealType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionIntegerToReal},
	{LESSER, RealType{}, RealType{}, BooleanType{}, ConversionNone, ConversionNone},
	{LESSER_EQ, IntegerType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionNone},
	{LESSER_EQ, IntegerType{}, RealType{}, BooleanType{}, ConversionIntegerToReal, ConversionNone},
	{LESSER_EQ, RealType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionIntegerToReal},
	{LESSER_EQ, RealType{}, RealType{}, BooleanType{}, ConversionNone, ConversionNone},

	// Equality is intentionally unavailable for strings.
	{EQUALS, IntegerType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionNone},
	{EQUALS, IntegerType{}, RealType{}, BooleanType{}, ConversionIntegerToReal, ConversionNone},
	{EQUALS, RealType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionIntegerToReal},
	{EQUALS, RealType{}, RealType{}, BooleanType{}, ConversionNone, ConversionNone},
	{EQUALS, BooleanType{}, BooleanType{}, BooleanType{}, ConversionNone, ConversionNone},
	{NOT_EQUALS, IntegerType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionNone},
	{NOT_EQUALS, IntegerType{}, RealType{}, BooleanType{}, ConversionIntegerToReal, ConversionNone},
	{NOT_EQUALS, RealType{}, IntegerType{}, BooleanType{}, ConversionNone, ConversionIntegerToReal},
	{NOT_EQUALS, RealType{}, RealType{}, BooleanType{}, ConversionNone, ConversionNone},
	{NOT_EQUALS, BooleanType{}, BooleanType{}, BooleanType{}, ConversionNone, ConversionNone},

	// Logical
	{AND, BooleanType{}, BooleanType{}, BooleanType{}, ConversionNone, ConversionNone},
	{OR, BooleanType{}, BooleanType{}, BooleanType{}, ConversionNone, ConversionNone},
}

var UnaryOperationTable = []UnaryOperationRule{
	{SUBTRACT, IntegerType{}, IntegerType{}, ConversionNone},
	{SUBTRACT, RealType{}, RealType{}, ConversionNone},
	{NOT, BooleanType{}, BooleanType{}, ConversionNone},
}

func CanCast(from Type, to Type) *ConversionKind {
	for _, cast := range AllowedCasts {
		if EqualTypes(cast.From, from) && EqualTypes(cast.To, to) {
			return &cast.ConversionKind
		}
	}
	return nil
}

func SearchConversion(with Type, c ConversionKind) *ImplicitCast {
	for _, cast := range AllowedCasts {
		if EqualTypes(cast.From, with) && cast.ConversionKind == c {
			return &cast
		}
	}
	return nil
}

func ResolveBinaryOperation(op OperatorType, left, right Type) (BinaryOperationRule, bool) {
	for _, rule := range BinaryOperationTable {
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
