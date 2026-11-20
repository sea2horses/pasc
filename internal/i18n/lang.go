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

	// Parser
	SyntaxErrorFound
	SyntaxErrorAtEOF
	FoundNewline
	FoundKeyword
	FoundString
	FoundToken
	InvalidSyntax
	ExpectedSymbol
	ExpectedNewline
	ExpectedTokenKind
	ExpectedKeyword
	ExpectedClosingKeyword
	ExpectedExpression
	ExpectedType
	ExpectedStatement
	ExpectedContinuation
	ExpectedSwitchCase
	ExpectedTimeUnit
	ExpectedItemAfterSeparator
	ExpectedEndOfProgram
	ExpectedDelimitedContent
	ExpectedClosingDelimiter

	IncorrectType
	VariableDoesntExist

	msgCount // sentinel, keep last
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

		// Parser
		SyntaxErrorFound:           "%[1]s, pero se encontró %[2]s",
		SyntaxErrorAtEOF:           "%[1]s, pero el archivo terminó",
		FoundNewline:               "un salto de línea",
		FoundKeyword:               "la palabra clave '%[1]s'",
		FoundString:                "el texto \"%[1]s\"",
		FoundToken:                 "'%[1]s'",
		InvalidSyntax:              "sintaxis inválida",
		ExpectedSymbol:             "se esperaba '%[1]s'",
		ExpectedNewline:            "se esperaba un salto de línea (cada instrucción va en su propia línea)",
		ExpectedTokenKind:          "se esperaba un token de tipo %[1]s",
		ExpectedKeyword:            "se esperaba la palabra clave '%[1]s'",
		ExpectedClosingKeyword:     "se esperaba '%[1]s' para cerrar el bloque '%[2]s'",
		ExpectedExpression:         "se esperaba una expresión (un valor, una variable o una operación)",
		ExpectedType:               "se esperaba un tipo de dato (Entero, Real, Caracter, Logico, etc.)",
		ExpectedStatement:          "se esperaba una instrucción",
		ExpectedContinuation:       "se esperaba la continuación de la instrucción",
		ExpectedSwitchCase:         "se esperaba un caso de Segun ('valor:' o 'De Otro Modo:')",
		ExpectedTimeUnit:           "se esperaba una unidad de tiempo (Segundos o Milisegundos)",
		ExpectedItemAfterSeparator: "se esperaba otro elemento después de '%[1]s'",
		ExpectedEndOfProgram:       "no se esperaba más código después del final del programa",
		ExpectedDelimitedContent:   "se esperaba contenido entre los delimitadores",
		ExpectedClosingDelimiter:   "falta el delimitador de cierre",

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

		// Parser
		SyntaxErrorFound:           "%[1]s, but found %[2]s",
		SyntaxErrorAtEOF:           "%[1]s, but the file ended",
		FoundNewline:               "a line break",
		FoundKeyword:               "the keyword '%[1]s'",
		FoundString:                "the text \"%[1]s\"",
		FoundToken:                 "'%[1]s'",
		InvalidSyntax:              "invalid syntax",
		ExpectedSymbol:             "expected '%[1]s'",
		ExpectedNewline:            "expected a line break (each statement goes on its own line)",
		ExpectedTokenKind:          "expected a token of kind %[1]s",
		ExpectedKeyword:            "expected the keyword '%[1]s'",
		ExpectedClosingKeyword:     "expected '%[1]s' to close the '%[2]s' block",
		ExpectedExpression:         "expected an expression (a value, a variable or an operation)",
		ExpectedType:               "expected a data type (Entero, Real, Caracter, Logico, etc.)",
		ExpectedStatement:          "expected a statement",
		ExpectedContinuation:       "expected the rest of the statement",
		ExpectedSwitchCase:         "expected a Segun case ('value:' or 'De Otro Modo:')",
		ExpectedTimeUnit:           "expected a time unit (Segundos or Milisegundos)",
		ExpectedItemAfterSeparator: "expected another item after '%[1]s'",
		ExpectedEndOfProgram:       "expected no more code after the end of the program",
		ExpectedDelimitedContent:   "expected content between the delimiters",
		ExpectedClosingDelimiter:   "missing closing delimiter",

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
