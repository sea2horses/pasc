package semantic

type PrimKind uint8

const (
	INVALID PrimKind = iota
	VOID
	UNTYPED_NUMBER
	INTEGER // translates to int64
	REAL    // translates to float64
	STRING  // translates to string
	BOOLEAN // translates to bool
)

type PrimitiveType interface {
	String() string
	Kind() PrimKind
	Signature() TypeSignature
}

type UntypedNumber_T struct{}

func (un UntypedNumber_T) Kind() PrimKind {
	return UNTYPED_NUMBER
}

func (un UntypedNumber_T) String() string {
	return "untyped_number"
}

func (un UntypedNumber_T) Signature() TypeSignature {
	return TypeSignature{BaseType: un.String()}
}

type Integer_T struct{}

func (i Integer_T) Kind() PrimKind {
	return INTEGER
}

func (i Integer_T) String() string {
	return "integer"
}

func (i Integer_T) Signature() TypeSignature {
	return TypeSignature{
		BaseType:      i.String(),
		GenericParams: nil,
	}
}

type Real_T struct{}

func (Real_T) Kind() PrimKind {
	return REAL
}

func (r Real_T) String() string {
	return "real"
}

func (r Real_T) Signature() TypeSignature {
	return TypeSignature{
		BaseType:      r.String(),
		GenericParams: nil,
	}
}

type String_T struct{}

func (s String_T) Kind() PrimKind {
	return STRING
}

func (s String_T) String() string {
	return "string"
}

func (s String_T) Signature() TypeSignature {
	return TypeSignature{
		BaseType:      s.String(),
		GenericParams: nil,
	}
}

type Bool_T struct{}

func (b Bool_T) Kind() PrimKind {
	return BOOLEAN
}

func (b Bool_T) String() string {
	return "boolean"
}

func (b Bool_T) Signature() TypeSignature {
	return TypeSignature{
		BaseType:      b.String(),
		GenericParams: nil,
	}
}

type Void_T struct{}

func (v Void_T) Kind() PrimKind {
	return VOID
}

func (v Void_T) String() string {
	return "void"
}

func (v Void_T) Signature() TypeSignature {
	return TypeSignature{
		BaseType:      v.String(),
		GenericParams: nil,
	}
}

type Invalid_T struct{}

func (i Invalid_T) Kind() PrimKind {
	return INVALID
}

func (i Invalid_T) String() string {
	return "<invalid>"
}

func (i Invalid_T) Signature() TypeSignature {
	return TypeSignature{
		BaseType:      i.String(),
		GenericParams: nil,
	}
}
