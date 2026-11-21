package i18n

import (
	"fmt"
	"strings"
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
	TypeInferredReal
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
	ImmutableConstant
	ImmutableFunction
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

	// Diagnostics, code generation & toolchain
	LevelError
	LevelWarning
	LevelInfo
	SummaryError
	SummaryErrors
	SummaryWarning
	SummaryWarnings
	SummaryNote
	SummaryNotes
	CompilationFailed
	CompilationFailedMany
	DiagnosticNotShown
	ReadSourceFailed
	FormatDiagnosticFailed
	WriteDiagnosticFailed
	SpanStartNotOneBased
	SpanEndNotOneBased
	SpanStartAfterEnd
	SpanStartOutsideFile
	SpanEndOutsideFile
	SpanColumnTooLarge
	CodegenUnsupportedNode
	CodegenUnmapped
	CodegenNonIntegralInteger
	CodegenNonNumericLiteral
	InternalCompilerError
	InvalidPath
	NotPscFile
	OpenFileFailed
	FileNotFound
	MainNotFound
	RuntimeExtMissing
	RuntimeModuleMissing
	RuntimeVendorBroken

	// CLI
	CmdRootShort
	CmdRootLong
	CmdBuildShort
	CmdBuildLong
	CmdRunShort
	CmdRunLong
	CmdHelpShort
	CmdHelpLong
	CmdCompletionShort
	CmdFileArg
	FlagDebug
	FlagFast
	FlagTemp
	FlagIndexBase
	FlagOutput
	FlagLang
	FlagHelp
	FlagVersion
	VersionWord
	UsageUsage
	UsageAliases
	UsageExamples
	UsageCommands
	UsageAdditionalCommands
	UsageFlags
	UsageGlobalFlags
	UsageHelpTopics
	UsageCommandPlaceholder
	UsageFlagsPlaceholder
	UsageDefault
	UsageMoreInfo
	ArgsExactlyOneFile
	InvalidLang

	// Formatter
	CmdFmtShort
	CmdFmtLong
	CmdFilesArg
	FlagWrite
	FlagCheck
	FlagIndent
	ArgsAtLeastOneFile
	InvalidIndent
	FmtSyntaxError
	FmtUnstable
	FmtNotFormatted
	FmtNotFormattedMany
	WriteFileFailed

	// Go toolchain setup
	GoSetupMissing
	GoSetupTooOld
	GoSetupOnce
	GoSetupDownloading
	GoSetupVerifying
	GoSetupExtracting
	GoSetupDone
	GoSetupInstalled
	GoSetupFirstBuild
	GoDownloadFailed
	GoChecksumMismatch
	GoExtractFailed
	GoManualHint

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
		TypeInferredReal:        "Real (inferido)",
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
		ImmutableConstant:       "'%[1]s' es una constante y no se puede modificar",
		ImmutableFunction:       "'%[1]s' es el nombre de una función y no se puede usar como variable",
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

		// Diagnostics, code generation & toolchain
		LevelError:                "error",
		LevelWarning:              "advertencia",
		LevelInfo:                 "nota",
		SummaryError:              "%[1]d error",
		SummaryErrors:             "%[1]d errores",
		SummaryWarning:            "%[1]d advertencia",
		SummaryWarnings:           "%[1]d advertencias",
		SummaryNote:               "%[1]d nota",
		SummaryNotes:              "%[1]d notas",
		CompilationFailed:         "no se pudo compilar el programa por %[1]d error",
		CompilationFailedMany:     "no se pudo compilar el programa por %[1]d errores",
		DiagnosticNotShown:        "<no se pudo mostrar el diagnóstico #%[1]d: %[2]s>",
		ReadSourceFailed:          "no se pudo leer el archivo fuente",
		FormatDiagnosticFailed:    "no se pudo dar formato al diagnóstico",
		WriteDiagnosticFailed:     "no se pudo escribir el diagnóstico",
		SpanStartNotOneBased:      "el inicio del rango debe empezar en 1 (se recibió línea %[1]d, columna %[2]d)",
		SpanEndNotOneBased:        "el final del rango debe empezar en 1 (se recibió línea %[1]d, columna %[2]d)",
		SpanStartAfterEnd:         "el inicio del rango está después de su final",
		SpanStartOutsideFile:      "la línea inicial %[1]d del rango está fuera del archivo",
		SpanEndOutsideFile:        "la línea final %[1]d del rango está fuera del archivo",
		SpanColumnTooLarge:        "la columna %[1]d excede el máximo de la línea %[2]d (%[3]d)",
		CodegenUnsupportedNode:    "nodo no soportado por el generador de código: %[1]T",
		CodegenUnmapped:           "%[1]s %[2]v no tiene representación en Go",
		CodegenNonIntegralInteger: "un literal con decimales se resolvió como Entero",
		CodegenNonNumericLiteral:  "un literal numérico tiene el tipo no numérico '%[1]s'",
		InternalCompilerError:     "error interno del compilador (esto es un bug de pascc, por favor repórtalo): %[1]v\n%[2]s",
		InvalidPath:               "la ruta está vacía o no es válida",
		NotPscFile:                "'%[1]s' no es un programa de PSeInt (el archivo debe terminar en .psc)",
		OpenFileFailed:            "no se pudo abrir '%[1]s'",
		FileNotFound:              "no existe el archivo '%[1]s'",
		MainNotFound:              "no se encontró el algoritmo principal",
		RuntimeExtMissing:         "falta la extensión embebida del runtime (pascc se compiló incorrectamente)",
		RuntimeModuleMissing:      "falta el módulo Go o las dependencias embebidas del runtime (pascc se compiló incorrectamente)",
		RuntimeVendorBroken:       "a las dependencias embebidas del runtime les falta vendor/modules.txt (pascc se compiló incorrectamente)",

		// CLI
		CmdRootShort:            "Compilador y analizador para el lenguaje PSeInt",
		CmdRootLong:             "Pasc analiza programas de PSeInt y los compila a ejecutables nativos usando el toolchain de Go.",
		CmdBuildShort:           "Compila un archivo .psc a un ejecutable nativo",
		CmdBuildLong:            "Analiza el programa en busca de errores y luego usa el toolchain de Go para convertirlo en un ejecutable nativo",
		CmdRunShort:             "Ejecuta un archivo .psc directamente",
		CmdRunLong:              "A diferencia de build, compila el programa, lo ejecuta y luego descarta el resultado",
		CmdHelpShort:            "Ayuda sobre cualquier comando",
		CmdHelpLong:             "Muestra la ayuda de cualquier comando.\nEscribe %[1]s help [comando] para ver todos los detalles.",
		CmdCompletionShort:      "Genera el script de autocompletado para la shell indicada",
		CmdFileArg:              "[archivo: *.psc]",
		FlagDebug:               "Muestra información de depuración sobre el proceso de compilación",
		FlagFast:                "Usa Enter para Esperar Tecla y omite las dependencias de terminal",
		FlagTemp:                "Compila el programa temporalmente y descarta el resultado (útil para pruebas)",
		FlagIndexBase:           "Primer índice de los arreglos: 0, 1 o auto (se deduce de cómo el programa indexa los arreglos)",
		FlagOutput:              "Carpeta donde se guarda el ejecutable",
		FlagLang:                "Idioma de los mensajes: es o en (por defecto se usa PASCC_LANG o el idioma del sistema)",
		FlagHelp:                "ayuda para %[1]s",
		FlagVersion:             "versión de %[1]s",
		VersionWord:             "versión",
		UsageUsage:              "Uso:",
		UsageAliases:            "Alias:",
		UsageExamples:           "Ejemplos:",
		UsageCommands:           "Comandos disponibles:",
		UsageAdditionalCommands: "Comandos adicionales:",
		UsageFlags:              "Opciones:",
		UsageGlobalFlags:        "Opciones globales:",
		UsageHelpTopics:         "Temas de ayuda adicionales:",
		UsageCommandPlaceholder: "[comando]",
		UsageFlagsPlaceholder:   "[opciones]",
		UsageDefault:            "por defecto",
		UsageMoreInfo:           "Usa \"%[1]s --help\" para más información sobre un comando.",
		ArgsExactlyOneFile:      "se esperaba exactamente un archivo .psc, pero se recibieron %[1]d argumento(s)",
		InvalidLang:             "idioma desconocido %[1]q (usa es o en)",

		// Formatter
		CmdFmtShort:         "Da formato a archivos .psc",
		CmdFmtLong:          "Reorganiza la sangría, los espacios y las mayúsculas de las palabras clave de programas PSeInt sin cambiar su significado. Conserva los comentarios y la estructura de líneas.\nSin opciones, muestra el resultado en la salida estándar. Las carpetas se recorren buscando archivos .psc.",
		CmdFilesArg:         "[archivos o carpetas...]",
		FlagWrite:           "Sobrescribe los archivos con el resultado en lugar de mostrarlo",
		FlagCheck:           "No modifica nada: lista los archivos sin formato y termina con error si hay alguno",
		FlagIndent:          "Sangría por nivel: tab o un número de espacios (1-8)",
		ArgsAtLeastOneFile:  "se esperaba al menos un archivo .psc o una carpeta",
		InvalidIndent:       "sangría no válida %[1]q (usa tab o un número de espacios entre 1 y 8)",
		FmtSyntaxError:      "no se dio formato a '%[1]s' porque tiene errores de sintaxis",
		FmtUnstable:         "el formateador generó un resultado distinto al programa original en '%[1]s' (esto es un bug de pascc, por favor repórtalo); el archivo no se modificó",
		FmtNotFormatted:     "%[1]d archivo necesita formato",
		FmtNotFormattedMany: "%[1]d archivos necesitan formato",
		WriteFileFailed:     "no se pudo escribir '%[1]s'",

		// Go toolchain setup
		GoSetupMissing:     "pascc necesita Go para convertir tus programas en ejecutables, y no lo encontró en este equipo.",
		GoSetupTooOld:      "pascc necesita Go %[2]s o más reciente para convertir tus programas en ejecutables, pero este equipo tiene Go %[1]s.",
		GoSetupOnce:        "Se descargará Go %[1]s (%[2]s) una sola vez. Las próximas compilaciones lo reutilizarán sin descargar nada.",
		GoSetupDownloading: "Descargando",
		GoSetupVerifying:   "Verificando la descarga...",
		GoSetupExtracting:  "Instalando...",
		GoSetupDone:        "listo",
		GoSetupInstalled:   "Go %[1]s quedó instalado en %[2]s",
		GoSetupFirstBuild:  "La primera compilación tarda un poco más mientras Go prepara su caché; las siguientes serán rápidas.",
		GoDownloadFailed:   "no se pudo descargar Go %[1]s: %[2]v",
		GoChecksumMismatch: "la descarga de Go está dañada (la suma SHA-256 no coincide); vuelve a intentarlo",
		GoExtractFailed:    "no se pudo instalar Go: %[1]v",
		GoManualHint:       "Si no tienes conexión a internet, instala Go manualmente desde https://go.dev/dl y vuelve a intentarlo.",
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
		TypeInferredReal:        "Real (inferred)",
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
		ImmutableConstant:       "'%[1]s' is a constant and can't be changed",
		ImmutableFunction:       "'%[1]s' is a function name and can't be used as a variable",
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

		// Diagnostics, code generation & toolchain
		LevelError:                "error",
		LevelWarning:              "warning",
		LevelInfo:                 "note",
		SummaryError:              "%[1]d error",
		SummaryErrors:             "%[1]d errors",
		SummaryWarning:            "%[1]d warning",
		SummaryWarnings:           "%[1]d warnings",
		SummaryNote:               "%[1]d note",
		SummaryNotes:              "%[1]d notes",
		CompilationFailed:         "the program could not be compiled because of %[1]d error",
		CompilationFailedMany:     "the program could not be compiled because of %[1]d errors",
		DiagnosticNotShown:        "<diagnostic #%[1]d could not be shown: %[2]s>",
		ReadSourceFailed:          "could not read the source file",
		FormatDiagnosticFailed:    "could not format the diagnostic",
		WriteDiagnosticFailed:     "could not write the diagnostic",
		SpanStartNotOneBased:      "span start must be 1-based (received line %[1]d, column %[2]d)",
		SpanEndNotOneBased:        "span end must be 1-based (received line %[1]d, column %[2]d)",
		SpanStartAfterEnd:         "span start comes after span end",
		SpanStartOutsideFile:      "span start line %[1]d is outside the source file",
		SpanEndOutsideFile:        "span end line %[1]d is outside the source file",
		SpanColumnTooLarge:        "column %[1]d exceeds the maximum of line %[2]d (%[3]d)",
		CodegenUnsupportedNode:    "node not supported by the code generator: %[1]T",
		CodegenUnmapped:           "%[1]s %[2]v has no Go representation",
		CodegenNonIntegralInteger: "a literal with decimals was resolved as Entero",
		CodegenNonNumericLiteral:  "a number literal has the non-numeric type '%[1]s'",
		InternalCompilerError:     "internal compiler error (this is a pascc bug, please report it): %[1]v\n%[2]s",
		InvalidPath:               "the path is empty or invalid",
		NotPscFile:                "'%[1]s' is not a PSeInt program (the file must end in .psc)",
		OpenFileFailed:            "could not open '%[1]s'",
		FileNotFound:              "the file '%[1]s' doesn't exist",
		MainNotFound:              "the main algorithm was not found",
		RuntimeExtMissing:         "the embedded runtime extension is missing (pascc was built incorrectly)",
		RuntimeModuleMissing:      "the runtime Go module or its embedded dependencies are missing (pascc was built incorrectly)",
		RuntimeVendorBroken:       "the embedded runtime dependencies are missing vendor/modules.txt (pascc was built incorrectly)",

		// CLI
		CmdRootShort:            "Compiler & analyzer tool for the PSeInt language",
		CmdRootLong:             "Pasc analyzes PSeInt programs and compiles them into native executables using the Go toolchain.",
		CmdBuildShort:           "Builds a .psc file into a native executable",
		CmdBuildLong:            "Analyzes the program for compile-time errors and then uses Go's toolchain to turn it into a native executable",
		CmdRunShort:             "Runs a .psc file, no fuss",
		CmdRunLong:              "Unlike build, this compiles the program, runs it and then discards the result",
		CmdHelpShort:            "Help about any command",
		CmdHelpLong:             "Help provides help for any command in the application.\nSimply type %[1]s help [command] for full details.",
		CmdCompletionShort:      "Generate the autocompletion script for the specified shell",
		CmdFileArg:              "[file: *.psc]",
		FlagDebug:               "Prints debug information about the compiling process",
		FlagFast:                "Uses an Enter prompt for Esperar Tecla and skips terminal dependencies",
		FlagTemp:                "Builds the program temporarily and then discards the result (useful for testing)",
		FlagIndexBase:           "First array index: 0, 1, or auto (deduced from how the program indexes arrays)",
		FlagOutput:              "Directory to write the executable to",
		FlagLang:                "Message language: es or en (defaults to PASCC_LANG or the system language)",
		FlagHelp:                "help for %[1]s",
		FlagVersion:             "version for %[1]s",
		VersionWord:             "version",
		UsageUsage:              "Usage:",
		UsageAliases:            "Aliases:",
		UsageExamples:           "Examples:",
		UsageCommands:           "Available Commands:",
		UsageAdditionalCommands: "Additional Commands:",
		UsageFlags:              "Flags:",
		UsageGlobalFlags:        "Global Flags:",
		UsageHelpTopics:         "Additional help topics:",
		UsageCommandPlaceholder: "[command]",
		UsageFlagsPlaceholder:   "[flags]",
		UsageDefault:            "default",
		UsageMoreInfo:           "Use \"%[1]s --help\" for more information about a command.",
		ArgsExactlyOneFile:      "expected exactly one .psc file, but got %[1]d argument(s)",
		InvalidLang:             "unknown language %[1]q (use es or en)",

		// Formatter
		CmdFmtShort:         "Formats .psc files",
		CmdFmtLong:          "Rewrites the indentation, spacing and keyword casing of PSeInt programs without changing their meaning. Comments and the line structure are kept.\nWithout flags, the result is printed to standard output. Directories are searched for .psc files.",
		CmdFilesArg:         "[files or directories...]",
		FlagWrite:           "Overwrite the files with the result instead of printing it",
		FlagCheck:           "Change nothing: list unformatted files and fail if there are any",
		FlagIndent:          "Indentation per level: tab or a number of spaces (1-8)",
		ArgsAtLeastOneFile:  "expected at least one .psc file or directory",
		InvalidIndent:       "invalid indentation %[1]q (use tab or a number of spaces between 1 and 8)",
		FmtSyntaxError:      "'%[1]s' was not formatted because it has syntax errors",
		FmtUnstable:         "the formatter produced output that differs from the original program in '%[1]s' (this is a pascc bug, please report it); the file was not modified",
		FmtNotFormatted:     "%[1]d file needs formatting",
		FmtNotFormattedMany: "%[1]d files need formatting",
		WriteFileFailed:     "could not write '%[1]s'",

		// Go toolchain setup
		GoSetupMissing:     "pascc needs Go to turn your programs into executables, and it isn't installed on this computer.",
		GoSetupTooOld:      "pascc needs Go %[2]s or newer to turn your programs into executables, but this computer has Go %[1]s.",
		GoSetupOnce:        "Go %[1]s (%[2]s) will be downloaded just this once. Future builds reuse it without downloading anything.",
		GoSetupDownloading: "Downloading",
		GoSetupVerifying:   "Verifying the download...",
		GoSetupExtracting:  "Installing...",
		GoSetupDone:        "done",
		GoSetupInstalled:   "Go %[1]s was installed in %[2]s",
		GoSetupFirstBuild:  "The first build takes a bit longer while Go warms up its cache; later builds are fast.",
		GoDownloadFailed:   "could not download Go %[1]s: %[2]v",
		GoChecksumMismatch: "the Go download is corrupted (SHA-256 mismatch); please try again",
		GoExtractFailed:    "could not install Go: %[1]v",
		GoManualHint:       "If you have no internet connection, install Go manually from https://go.dev/dl and try again.",
	},
}

var current = ES

func Set(l Lang) { current = l }

// Parse reads a language code or a POSIX locale ("es", "en", "es_NI.UTF-8").
func Parse(s string) (Lang, bool) {
	code := strings.ToLower(s)
	if i := strings.IndexAny(code, "_-.@"); i >= 0 {
		code = code[:i]
	}
	switch code {
	case "es":
		return ES, true
	case "en":
		return EN, true
	}
	return ES, false
}

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
