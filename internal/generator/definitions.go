package generator

import "github.com/dave/jennifer/jen"

// Jennifer statements are mutable. Keep constructors here instead of shared
// statements so each generated call owns its syntax tree.
func PowerFunc() *jen.Statement           { return jen.Qual("math", "Pow") }
func SinFunc() *jen.Statement             { return jen.Qual("math", "Sin") }
func CosFunc() *jen.Statement             { return jen.Qual("math", "Cos") }
func TruncFunc() *jen.Statement           { return jen.Qual("math", "Trunc") }
func WriteFunc() *jen.Statement           { return jen.Id("Write") }
func WriteContinuousFunc() *jen.Statement { return jen.Id("WriteContinuous") }
func ReadFunc() *jen.Statement            { return jen.Id("Read") }
func ToIntegerExFunc() *jen.Statement     { return jen.Id("ToIntegerExact") }
func IntToRealFunc() *jen.Statement       { return jen.Id("IntegerToReal") }
func MakeTensorFunc() *jen.Statement      { return jen.Id("NewTensor") }
func IndexTensor() *jen.Statement         { return jen.Id("IndexTensor") }
func IterFunc() *jen.Statement            { return jen.Id("MakeIter") }
func IterSteplessFunc() *jen.Statement    { return jen.Id("MakeIterStepless") }
func ClearScreenFunc() *jen.Statement     { return jen.Id("ClearScreen") }
func TimeoutFunc() *jen.Statement         { return jen.Id("Timeout") }
