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
	UnterminatedStringMismatch
	UnterminatedStringMissingClose

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

	// Analyzer
	IncorrectType
	VariableDoesntExist
	TypeArray
	TypeArrayND
	TypeVoid
	TypeFunction
	TypeUnknown
	UnaryOpNotAllowed
	BinaryOpNotAllowed
	SymbolAlreadyExists
	UnsupportedWriteType
	ExpectedValueOfType
	UnexpectedVoid
	NotAssignable
	Immutable
	DeclaredHere
	DeclaredHereWithType
	ImplicitDeclaration
	UnknownType
	DimensionNotVariable
	AlreadyDimensioned
	DimensionMustBeInteger
	DimensionAtLeastOne
	RedimNotDimensioned
	RedimNotArray
	RedimRankChange
	IndexNotVariable
	IndexNotArray
	IndexCountMismatch
	IndexMustBeInteger
	IndexOutOfBounds
	IndexMaybeOutOfBounds
	MixedIndexBase
	InvalidIndexBase
	RealToInteger
	RealToIntegerRuntime
	FunctionAlreadyDeclared
	DuplicateParameter
	IncompatibleReturn
	CannotInferParameter
	CannotInferReturn
	NotCallable
	NotCallableOrIndexable
	ArgCountMismatch
	ArgMustBeVariable
	ArgTypeMismatch
	CaseTypeMismatch
	ExpressionStatement
	UnsupportedStatement
	UnsupportedExpression
	WaitAmountNumeric
	ForNeedsVariable
	ForEachNeedsArray

	msgCount // sentinel, keep last
)

var catalog = map[Lang]map[Msg]string{
	ES: {
		// Lexer
		UnexpectedChar:                 "carácter inesperado: '%[1]c'",
		CharOutOfBounds:                "carácter (%[1]d) fuera de rango (máximo %[2]d)",
		ExpectedIdentifier:             "se esperaba un identificador (nombre de variable, función, etc.)",
		ExpectedString:                 "se esperaba una cadena/string",
		ExpectedNumber:                 "se esperaba un número",
		UnterminatedString:             "string sin terminar (todos los strings deben tener un \" de apertura y de cierre)",
		UnterminatedBlockComment:       "comentario multi-linea sin terminar (cerrar con */)",
		UnterminatedStringMismatch:     "string sin terminar: se abrió con %[1]c pero la línea lo cierra con una comilla distinta (usa la misma comilla para abrir y cerrar)",
		UnterminatedStringMissingClose: "string sin terminar: falta el %[1]c de cierre en esta línea",

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

		// Analyzer
		IncorrectType:           "tipo incorrecto: se esperaba '%[1]s', pero se obtuvo '%[2]s'",
		VariableDoesntExist:     "la variable '%[1]s' no existe",
		TypeArray:               "arreglo de %[1]s",
		TypeArrayND:             "arreglo de %[1]s de %[2]d dimensiones",
		TypeVoid:                "nada",
		TypeFunction:            "Funcion",
		TypeUnknown:             "desconocido",
		UnaryOpNotAllowed:       "el operador '%[1]s' no se puede usar con un valor de tipo '%[2]s'",
		BinaryOpNotAllowed:      "el operador '%[1]s' no se puede usar entre '%[2]s' y '%[3]s'",
		SymbolAlreadyExists:     "'%[1]s' ya está definido",
		UnsupportedWriteType:    "no se puede mostrar un valor de tipo '%[1]s' (solo Entero, Real, Cadena o Logico)",
		ExpectedValueOfType:     "se esperaba un valor de tipo '%[1]s'",
		UnexpectedVoid:          "esta expresión no devuelve ningún valor (¿es una llamada a un SubProceso?)",
		NotAssignable:           "aquí no se puede guardar un valor (solo en variables o posiciones de un arreglo)",
		Immutable:               "este valor no se puede modificar (es una constante o el nombre de una función)",
		DeclaredHere:            "declarado aquí",
		DeclaredHereWithType:    "'%[1]s' se declaró aquí con tipo '%[2]s'",
		ImplicitDeclaration:     "la variable '%[1]s' no fue declarada con Definir; se declarará automáticamente",
		UnknownType:             "tipo de dato desconocido '%[1]s' (usa Entero, Real, Caracter o Logico)",
		DimensionNotVariable:    "no se puede dimensionar '%[1]s' porque no es una variable",
		AlreadyDimensioned:      "'%[1]s' ya fue dimensionado (usa Redimensionar para cambiar su tamaño)",
		DimensionMustBeInteger:  "el tamaño de una dimensión debe ser un número entero",
		DimensionAtLeastOne:     "el tamaño de una dimensión debe ser al menos 1, pero es %[1]s",
		RedimNotDimensioned:     "no se puede redimensionar '%[1]s' porque no ha sido dimensionado (usa Dimensionar primero)",
		RedimNotArray:           "no se puede redimensionar '%[1]s' porque no es un arreglo",
		RedimRankChange:         "no se puede cambiar la cantidad de dimensiones de '%[1]s' de %[2]d a %[3]d",
		IndexNotVariable:        "solo se pueden indexar variables",
		IndexNotArray:           "no se puede usar [] en un valor de tipo '%[1]s' (no es un arreglo)",
		IndexCountMismatch:      "cantidad de índices incorrecta: el arreglo tiene %[1]d dimensión(es), pero se usaron %[2]d índice(s)",
		IndexMustBeInteger:      "los índices de un arreglo deben ser números enteros",
		IndexOutOfBounds:        "el índice %[1]s está fuera de rango para '%[2]s' (los índices válidos van de %[3]s a %[4]s en arreglos de base %[5]s)",
		IndexMaybeOutOfBounds:   "el índice puede valer de %[1]s a %[2]s, y se sale del rango de '%[3]s' (los índices válidos van de %[4]s a %[5]s en arreglos de base %[6]s)",
		MixedIndexBase:          "este acceso parece usar base %[1]s, pero la mayor parte del programa indexa los arreglos en base %[2]s; se asumirá base %[2]s (usa --index-base para elegir)",
		InvalidIndexBase:        "base de índices inválida %[1]q (se esperaba auto, 0 o 1)",
		RealToInteger:           "el valor real %[1]s no se puede convertir a entero; usa Redon() o Trunc()",
		RealToIntegerRuntime:    "convertir un real a entero falla al ejecutar si el valor no es exacto; usa Redon() o Trunc()",
		FunctionAlreadyDeclared: "la función '%[1]s' ya fue declarada",
		DuplicateParameter:      "el parámetro '%[1]s' aparece más de una vez",
		IncompatibleReturn:      "la función '%[1]s' devuelve un valor de tipo incompatible",
		CannotInferParameter:    "no se pudo deducir el tipo del parámetro '%[1]s'; se usará Cadena",
		CannotInferReturn:       "no se pudo deducir el tipo de retorno de la función '%[1]s'; se usará Cadena",
		NotCallable:             "un valor de tipo '%[1]s' no se puede llamar como función",
		NotCallableOrIndexable:  "un valor de tipo '%[1]s' no se puede llamar ni indexar",
		ArgCountMismatch:        "se esperaban %[1]d argumento(s), pero se pasaron %[2]d",
		ArgMustBeVariable:       "el argumento %[1]d debe ser una variable modificable, porque el parámetro es Por Referencia",
		ArgTypeMismatch:         "argumento %[1]d: se esperaba '%[2]s', pero se obtuvo '%[3]s'",
		CaseTypeMismatch:        "este caso es de tipo '%[1]s' y no se puede comparar con el valor del Segun, que es de tipo '%[2]s'",
		ExpressionStatement:     "esta expresión no hace nada por sí sola; solo las llamadas a funciones pueden usarse como instrucción",
		UnsupportedStatement:    "el compilador aún no soporta esta instrucción (%[1]T)",
		UnsupportedExpression:   "expresión inválida o aún no soportada",
		WaitAmountNumeric:       "el tiempo de espera debe ser un número",
		ForNeedsVariable:        "el contador de un Para debe ser una variable",
		ForEachNeedsArray:       "'Para Cada' necesita un arreglo, pero se obtuvo '%[1]s'",
	},
	EN: {
		// Lexer
		UnexpectedChar:                 "unexpected character: '%[1]c'",
		CharOutOfBounds:                "character (%[1]d) out of bounds (max. %[2]d)",
		ExpectedIdentifier:             "expected an identifier (name of a variable, function, etc.)",
		ExpectedString:                 "expected a string",
		ExpectedNumber:                 "expected a number",
		UnterminatedString:             "unterminated string (all strings must have a \" character at both their start and end)",
		UnterminatedBlockComment:       "unterminated multi-line comment (close with */)",
		UnterminatedStringMismatch:     "unterminated string: opened with %[1]c but the line closes it with a different quote (use the same quote to open and close)",
		UnterminatedStringMissingClose: "unterminated string: missing closing %[1]c on this line",

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

		// Analyzer
		IncorrectType:           "incorrect type: expected '%[1]s', but got '%[2]s'",
		VariableDoesntExist:     "variable '%[1]s' doesn't exist",
		TypeArray:               "array of %[1]s",
		TypeArrayND:             "%[2]d-dimensional array of %[1]s",
		TypeVoid:                "nothing",
		TypeFunction:            "Funcion",
		TypeUnknown:             "unknown",
		UnaryOpNotAllowed:       "the '%[1]s' operator can't be used on a value of type '%[2]s'",
		BinaryOpNotAllowed:      "the '%[1]s' operator can't be used between '%[2]s' and '%[3]s'",
		SymbolAlreadyExists:     "'%[1]s' is already defined",
		UnsupportedWriteType:    "a value of type '%[1]s' can't be displayed (only Entero, Real, Cadena or Logico)",
		ExpectedValueOfType:     "expected a value of type '%[1]s'",
		UnexpectedVoid:          "this expression doesn't return a value (is it a call to a SubProceso?)",
		NotAssignable:           "a value can't be stored here (only in variables or array elements)",
		Immutable:               "this value can't be changed (it's a constant or a function name)",
		DeclaredHere:            "declared here",
		DeclaredHereWithType:    "'%[1]s' was declared here with type '%[2]s'",
		ImplicitDeclaration:     "variable '%[1]s' wasn't declared with Definir; declaring it automatically",
		UnknownType:             "unknown data type '%[1]s' (use Entero, Real, Caracter or Logico)",
		DimensionNotVariable:    "'%[1]s' can't be dimensioned because it isn't a variable",
		AlreadyDimensioned:      "'%[1]s' is already dimensioned (use Redimensionar to change its size)",
		DimensionMustBeInteger:  "dimension sizes must be integers",
		DimensionAtLeastOne:     "dimension sizes must be at least 1, got %[1]s",
		RedimNotDimensioned:     "'%[1]s' can't be redimensioned because it hasn't been dimensioned (use Dimensionar first)",
		RedimNotArray:           "'%[1]s' can't be redimensioned because it isn't an array",
		RedimRankChange:         "can't change the number of dimensions of '%[1]s' from %[2]d to %[3]d",
		IndexNotVariable:        "only variables can be indexed",
		IndexNotArray:           "can't use [] on a value of type '%[1]s' (it isn't an array)",
		IndexCountMismatch:      "wrong number of indexes: the array has %[1]d dimension(s), but %[2]d index(es) were given",
		IndexMustBeInteger:      "array indexes must be integers",
		IndexOutOfBounds:        "index %[1]s is out of bounds for '%[2]s' (valid indexes are %[3]s to %[4]s with %[5]s-based arrays)",
		IndexMaybeOutOfBounds:   "the index can range from %[1]s to %[2]s, which goes out of bounds for '%[3]s' (valid indexes are %[4]s to %[5]s with %[6]s-based arrays)",
		MixedIndexBase:          "this access looks %[1]s-based, but most of the program indexes arrays %[2]s-based; assuming %[2]s-based (pass --index-base to choose)",
		InvalidIndexBase:        "invalid index base %[1]q (expected auto, 0 or 1)",
		RealToInteger:           "the real value %[1]s can't be converted to an integer; use Redon() or Trunc()",
		RealToIntegerRuntime:    "converting a real to an integer fails at runtime if the value isn't whole; use Redon() or Trunc()",
		FunctionAlreadyDeclared: "function '%[1]s' is already declared",
		DuplicateParameter:      "parameter '%[1]s' appears more than once",
		IncompatibleReturn:      "function '%[1]s' returns a value of an incompatible type",
		CannotInferParameter:    "couldn't infer the type of parameter '%[1]s'; assuming Cadena",
		CannotInferReturn:       "couldn't infer the return type of function '%[1]s'; assuming Cadena",
		NotCallable:             "a value of type '%[1]s' can't be called as a function",
		NotCallableOrIndexable:  "a value of type '%[1]s' can't be called or indexed",
		ArgCountMismatch:        "expected %[1]d argument(s), but %[2]d were given",
		ArgMustBeVariable:       "argument %[1]d must be a modifiable variable, because the parameter is Por Referencia",
		ArgTypeMismatch:         "argument %[1]d: expected '%[2]s', but got '%[3]s'",
		CaseTypeMismatch:        "this case has type '%[1]s' and can't be compared with the Segun value, which has type '%[2]s'",
		ExpressionStatement:     "this expression does nothing on its own; only function calls can be used as statements",
		UnsupportedStatement:    "the compiler doesn't support this statement yet (%[1]T)",
		UnsupportedExpression:   "invalid or not yet supported expression",
		WaitAmountNumeric:       "the wait time must be a number",
		ForNeedsVariable:        "the counter of a Para loop must be a variable",
		ForEachNeedsArray:       "'Para Cada' needs an array, but got '%[1]s'",
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
