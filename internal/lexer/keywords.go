package lexer

import "strings"

//go:generate stringer -type=Keyword
type Keyword int8

const (
	// Main
	ALGORITMO Keyword = iota
	FINALGORITMO

	PROCESO
	FINPROCESO

	// Que
	QUE

	// i/o
	ESCRIBIR
	IMPRIMIR
	LEER

	// Operators
	MOD

	// Flow Control
	SI
	ENTONCES
	SINO
	FINSI

	// Switch
	SEGUN
	FINSEGUN
	DE
	OTRO
	MODO

	// Declaracion / Tipos
	DEFINIR
	COMO
	ENTERO
	CADENA
	TEXTO
	CARACTER
	LOGICO
	REAL
	NUMERO
	NUMERICO

	// Arreglos
	DIMENSIONAR
	DIMENSION
	REDIMENSIONAR

	MIENTRAS
	HACER
	FINMIENTRAS

	PARA
	HASTA
	CON
	PASO
	FINPARA

	REPETIR

	SIN
	SALTAR

	// Extra
	BORRAR
	LIMPIAR
	PANTALLA

	ESPERAR
	TECLA
	SEGUNDOS
	MILISEGUNDOS
	FIN

	Y
	O
	CASO
	FUNCION
	FINFUNCION
)

func MapToKeyword(str []rune) (Keyword, bool) {
	val, ok := keywordMap[strings.ToLower(string(str))]
	return val, ok
}

var keywordMap map[string]Keyword = map[string]Keyword{
	"algoritmo":     ALGORITMO,
	"finalgoritmo":  FINALGORITMO,
	"proceso":       PROCESO,
	"finproceso":    FINPROCESO,
	"escribir":      ESCRIBIR,
	"imprimir":      IMPRIMIR,
	"leer":          LEER,
	"mientras":      MIENTRAS,
	"hacer":         HACER,
	"finmientras":   FINMIENTRAS,
	"mod":           MOD,
	"si":            SI,
	"entonces":      ENTONCES,
	"sino":          SINO,
	"finsi":         FINSI,
	"definir":       DEFINIR,
	"como":          COMO,
	"borrar":        BORRAR,
	"limpiar":       LIMPIAR,
	"pantalla":      PANTALLA,
	"segun":         SEGUN,
	"caso":          CASO,
	"finsegun":      FINSEGUN,
	"de":            DE,
	"otro":          OTRO,
	"modo":          MODO,
	"dimensionar":   DIMENSIONAR,
	"dimension":     DIMENSION,
	"redimensionar": REDIMENSIONAR,
	"para":          PARA,
	"finpara":       FINPARA,
	"repetir":       REPETIR,
	"hasta":         HASTA,
	"con":           CON,
	"paso":          PASO,
	"sin":           SIN,
	"saltar":        SALTAR,
	"esperar":       ESPERAR,
	"tecla":         TECLA,
	"segundos":      SEGUNDOS,
	"milisegundos":  MILISEGUNDOS,
	"fin":           FIN,
	"funcion":       FUNCION,
	"función":       FUNCION,
	"finfuncion":    FINFUNCION,
	"finfunción":    FINFUNCION,
	"que":           QUE,
	"y":             Y,
	"o":             O,
}
