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

func CanCast(from Type, to Type) *ConversionKind {
	from, to = Resolve(from), Resolve(to)
	for _, cast := range allowedCasts {
		if EqualTypes(cast.From, from) && EqualTypes(cast.To, to) {
			kind := cast.ConversionKind
			return &kind
		}
	}
	return nil
}

func SearchConversion(with Type, kind ConversionKind) *ImplicitCast {
	with = Resolve(with)
	for _, cast := range allowedCasts {
		if EqualTypes(cast.From, with) && cast.ConversionKind == kind {
			result := cast
			return &result
		}
	}
	return nil
}

// operationType returns the known numeric type, or false when the operand is
// an unresolved inference variable/untyped literal.
func operationType(t Type) (Type, bool) {
	t = Resolve(t)
	if infer, ok := t.(*InferType); ok && infer.Resolved == nil {
		return nil, false
	}
	if EqualTypes(t, UntypedNumber_T{}) {
		return nil, false
	}
	if IsInteger(t) || IsReal(t) {
		return t, true
	}
	return nil, false
}

func isUnknownNumeric(t Type) bool {
	_, known := operationType(t)
	return !known && (IsNumeric(Resolve(t)) || unresolvedInference(t))
}

func unresolvedInference(t Type) bool {
	infer, ok := Resolve(t).(*InferType)
	return ok && infer.Resolved == nil
}

func conversionFor(source, target Type) ConversionKind {
	if EqualTypes(source, target) {
		return ConversionNone
	}
	if kind := CanCast(source, target); kind != nil {
		return *kind
	}
	return ConversionNone
}

func numericOperands(left, right Type, forceReal bool) (Type, Type, Type, ConversionKind, ConversionKind, bool) {
	leftType, leftKnown := operationType(left)
	rightType, rightKnown := operationType(right)
	if (!leftKnown && !isUnknownNumeric(left)) || (!rightKnown && !isUnknownNumeric(right)) {
		return nil, nil, nil, ConversionNone, ConversionNone, false
	}

	// An unresolved number follows its known peer. With no peer, ordinary
	// arithmetic defaults to integer; division and power need real operands.
	var common Type = Integer_T{}
	if forceReal || (leftKnown && IsReal(leftType)) || (rightKnown && IsReal(rightType)) {
		common = Real_T{}
	}
	if !leftKnown {
		leftType = common
	}
	if !rightKnown {
		rightType = common
	}
	leftConversion := conversionFor(leftType, common)
	rightConversion := conversionFor(rightType, common)
	return leftType, rightType, common, leftConversion, rightConversion, true
}

func ResolveBinaryOperation(op OperatorType, left, right Type) (BinaryOperationRule, bool) {
	rule := BinaryOperationRule{Operator: op}

	switch op {
	case ADD:
		if EqualTypes(Resolve(left), String_T{}) && EqualTypes(Resolve(right), String_T{}) {
			return BinaryOperationRule{Operator: op, Left: String_T{}, Right: String_T{}, Result: String_T{}}, true
		}
		fallthrough
	case SUBTRACT, MULTIPLY:
		l, r, common, lc, rc, ok := numericOperands(left, right, false)
		if !ok {
			return rule, false
		}
		return BinaryOperationRule{Operator: op, Left: l, Right: r, Result: common, LeftConversion: lc, RightConversion: rc}, true

	case DIVIDE, POWER:
		l, r, _, lc, rc, ok := numericOperands(left, right, true)
		if !ok {
			return rule, false
		}
		return BinaryOperationRule{Operator: op, Left: l, Right: r, Result: Real_T{}, LeftConversion: lc, RightConversion: rc}, true

	case MODULO:
		l, leftKnown := operationType(left)
		r, rightKnown := operationType(right)
		if (!leftKnown && !isUnknownNumeric(left)) || (!rightKnown && !isUnknownNumeric(right)) {
			return rule, false
		}
		if !leftKnown {
			l = Integer_T{}
		}
		if !rightKnown {
			r = Integer_T{}
		}
		switch {
		case IsInteger(l) && IsInteger(r):
			return BinaryOperationRule{Operator: op, Left: l, Right: r, Result: Integer_T{}}, true
		case IsInteger(l) && IsReal(r):
			return BinaryOperationRule{Operator: op, Left: l, Right: r, Result: Integer_T{}, RightConversion: ConversionRealToIntegerExact}, true
		case IsReal(l) && IsInteger(r):
			return BinaryOperationRule{Operator: op, Left: l, Right: r, Result: Real_T{}, LeftConversion: ConversionRealToIntegerExact}, true
		case IsReal(l) && IsReal(r):
			return BinaryOperationRule{Operator: op, Left: l, Right: r, Result: Integer_T{}, LeftConversion: ConversionRealToIntegerExact, RightConversion: ConversionRealToIntegerExact}, true
		default:
			return rule, false
		}

	case GREATER, GREATER_EQ, LESSER, LESSER_EQ, EQUALS, NOT_EQUALS:
		if (op == EQUALS || op == NOT_EQUALS) && EqualTypes(Resolve(left), Bool_T{}) && EqualTypes(Resolve(right), Bool_T{}) {
			return BinaryOperationRule{Operator: op, Left: Bool_T{}, Right: Bool_T{}, Result: Bool_T{}}, true
		}
		l, r, common, lc, rc, ok := numericOperands(left, right, false)
		if !ok {
			return rule, false
		}
		if IsReal(common) {
			lc = conversionFor(l, Real_T{})
			rc = conversionFor(r, Real_T{})
		}
		return BinaryOperationRule{Operator: op, Left: l, Right: r, Result: Bool_T{}, LeftConversion: lc, RightConversion: rc}, true

	case AND, OR:
		if EqualTypes(Resolve(left), Bool_T{}) && EqualTypes(Resolve(right), Bool_T{}) {
			return BinaryOperationRule{Operator: op, Left: Bool_T{}, Right: Bool_T{}, Result: Bool_T{}}, true
		}
	}

	return rule, false
}

func ResolveUnaryOperation(op OperatorType, operand Type) (UnaryOperationRule, bool) {
	switch op {
	case SUBTRACT:
		if t, known := operationType(operand); known {
			return UnaryOperationRule{Operator: op, Operand: t, Result: t}, true
		}
		if isUnknownNumeric(operand) {
			return UnaryOperationRule{Operator: op, Operand: Integer_T{}, Result: Integer_T{}}, true
		}
	case NOT:
		if EqualTypes(Resolve(operand), Bool_T{}) {
			return UnaryOperationRule{Operator: op, Operand: Bool_T{}, Result: Bool_T{}}, true
		}
	}
	return UnaryOperationRule{}, false
}
