package lexer

import "strings"

//go:generate stringer -type=Keyword
type Keyword int8

const (
	// Main
	ALGORITMO Keyword = iota
	FINALGORITMO

	// Que
	QUE

	// i/o
	ESCRIBIR
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
	LOGICO
	REAL

	// Arreglos
	DIMENSIONAR
	REDIMENSIONAR

	MIENTRAS
	HACER
	FINMIENTRAS

	PARA
	HASTA
	CON
	PASO
	FINPARA

	SIN
	SALTAR

	// Extra
	BORRAR
	PANTALLA

	ESPERAR
	SEGUNDOS
	MILISEGUNDOS
)

func MapToKeyword(str []rune) (Keyword, bool) {
	val, ok := keywordMap[strings.ToLower(string(str))]
	return val, ok
}

var keywordMap map[string]Keyword = map[string]Keyword{
	"algoritmo":     ALGORITMO,
	"finalgoritmo":  FINALGORITMO,
	"escribir":      ESCRIBIR,
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
	"entero":        ENTERO,
	"cadena":        CADENA,
	"logico":        LOGICO,
	"real":          REAL,
	"borrar":        BORRAR,
	"pantalla":      PANTALLA,
	"segun":         SEGUN,
	"finsegun":      FINSEGUN,
	"de":            DE,
	"otro":          OTRO,
	"modo":          MODO,
	"dimensionar":   DIMENSIONAR,
	"redimensionar": REDIMENSIONAR,
	"para":          PARA,
	"finpara":       FINPARA,
	"hasta":         HASTA,
	"con":           CON,
	"paso":          PASO,
	"sin":           SIN,
	"saltar":        SALTAR,
	"esperar":       ESPERAR,
	"segundos":      SEGUNDOS,
	"milisegundos":  MILISEGUNDOS,
	"que":           QUE,
}
