package i18n

import (
	"fmt"
)

type Lang uint8

const (
	ES Lang = iota
	EN
)

type Msg uint16

//go:generate stringer -type=Msg
const (
	// Lexer
	UnexpectedChar Msg = iota
	CharOutOfBounds

	ExpectedIdentifier
	ExpectedString
	ExpectedNumber
	UnterminatedString
	UnterminatedBlockComment

	IncorrectType
	VariableDoesntExist
)

var catalog = map[Lang]map[Msg]string{
	ES: {
		// Lexer
		UnexpectedChar:           "carácter inesperado: '%[1]s'",
		CharOutOfBounds:          "carácter (%[1]s) fuera de rango (máximo. %[2]s)",
		ExpectedIdentifier:       "se esperaba un identificador (nombre de variable, función, etc.)",
		ExpectedString:           "se esperaba una cadena/string",
		ExpectedNumber:           "se esperaba un número",
		UnterminatedString:       "string sin terminar (todos los strings deben tener un \" de apertura y de cierre)",
		UnterminatedBlockComment: "comentario multi-linea sin terminar (cerrar con */)",

		IncorrectType:       "tipo incorrecto: se esperaba %[1]s, se obtuvo %[2]s",
		VariableDoesntExist: "la variable '%[1]s' no existe",
	},
	EN: {
		// Lexer
		UnexpectedChar:           "unexpected character: '%[1]s'",
		CharOutOfBounds:          "char (%[1]s) out of bounds (max. %[2]s)",
		ExpectedIdentifier:       "expected an identifier (name of a variable, function, etc.)",
		ExpectedString:           "expected a string",
		ExpectedNumber:           "expected a number",
		UnterminatedString:       "unterminated string (all strings must have a \" character at both their start and end",
		UnterminatedBlockComment: "unterminated multi-line comment (close with */)",

		IncorrectType:       "incorrect type: expected %[1]s, got %[2]s",
		VariableDoesntExist: "variable '%[1]s' doesn't exist",
	},
}

var current = ES

func Set(l Lang) { current = l }

func T(m Msg, args ...any) string {
	f, ok := catalog[current][m]
	// english fallback
	if !ok {
		f = catalog[EN][m]
	}
	return fmt.Sprintf(f, args...)
}

// For non-constant generation
func TFunc(m Msg, args ...any) func() string {
	return func() string {
		return T(m, args...)
	}
}
