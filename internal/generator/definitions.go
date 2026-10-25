package generator

import "github.com/dave/jennifer/jen"

var (
	PowerFunc *jen.Statement = jen.Qual("math", "Pow")
	SinFunc   *jen.Statement = jen.Qual("math", "Sin")
	CosFunc   *jen.Statement = jen.Qual("math", "Sin")

	/* I/O */
	WriteFunc           *jen.Statement = jen.Id("Write")
	WriteContinuousFunc *jen.Statement = jen.Id("WriteContinuous")
	ReadFunc            *jen.Statement = jen.Id("Read")

	/* Conversions */
	ToIntegerExFunc *jen.Statement = jen.Id("ToIntegerExact")
	IntToRealFunc   *jen.Statement = jen.Id("IntegerToReal")

	/* Utils */
	ClearScreen *jen.Statement = jen.Id("ClearScreen")
	Timeout     *jen.Statement = jen.Id("Timeout")
)
