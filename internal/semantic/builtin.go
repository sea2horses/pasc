package semantic

type BuiltinKind uint8

const (
	BuiltinSqrt BuiltinKind = iota
	BuiltinSin
	BuiltinCos
	BuiltinTrunc
	BuiltinAbs
	BuiltinLn
	BuiltinExp
	BuiltinTan
	BuiltinASin
	BuiltinACos
	BuiltinATan
	BuiltinRound
	BuiltinRandom
	BuiltinRandomRange
	BuiltinLength
	BuiltinUpper
	BuiltinLower
	BuiltinSubstring
	BuiltinConcat
	BuiltinToNumber
	BuiltinToText
	BuiltinCurrentTime
	BuiltinCurrentDate
)

type BuiltinInfo struct {
	Kind BuiltinKind
}
