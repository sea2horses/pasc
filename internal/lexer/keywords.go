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
	NO
	CASO
	FUNCION
	FINFUNCION
	POR
	REFERENCIA
	SUBPROCESO
	FINSUBPROCESO
)

func MapToKeyword(str []rune) (Keyword, bool) {
	val, ok := keywordMap[strings.ToLower(string(str))]
	return val, ok
}

var keywordMap map[string]Keyword = map[string]Keyword{
	"algoritmo":       ALGORITMO,
	"finalgoritmo":    FINALGORITMO,
	"proceso":         PROCESO,
	"finproceso":      FINPROCESO,
	"escribir":        ESCRIBIR,
	"imprimir":        IMPRIMIR,
	"leer":            LEER,
	"mientras":        MIENTRAS,
	"hacer":           HACER,
	"finmientras":     FINMIENTRAS,
	"mod":             MOD,
	"si":              SI,
	"entonces":        ENTONCES,
	"sino":            SINO,
	"finsi":           FINSI,
	"definir":         DEFINIR,
	"como":            COMO,
	"borrar":          BORRAR,
	"limpiar":         LIMPIAR,
	"pantalla":        PANTALLA,
	"segun":           SEGUN,
	"caso":            CASO,
	"finsegun":        FINSEGUN,
	"de":              DE,
	"otro":            OTRO,
	"modo":            MODO,
	"dimensionar":     DIMENSIONAR,
	"dimension":       DIMENSION,
	"redimensionar":   REDIMENSIONAR,
	"para":            PARA,
	"finpara":         FINPARA,
	"repetir":         REPETIR,
	"hasta":           HASTA,
	"con":             CON,
	"paso":            PASO,
	"sin":             SIN,
	"saltar":          SALTAR,
	"esperar":         ESPERAR,
	"tecla":           TECLA,
	"segundos":        SEGUNDOS,
	"milisegundos":    MILISEGUNDOS,
	"fin":             FIN,
	"funcion":         FUNCION,
	"función":         FUNCION,
	"finfuncion":      FINFUNCION,
	"finfunción":      FINFUNCION,
	"por":             POR,
	"referencia":      REFERENCIA,
	"subproceso":      SUBPROCESO,
	"subalgoritmo":    SUBPROCESO,
	"finsubproceso":   FINSUBPROCESO,
	"finsubalgoritmo": FINSUBPROCESO,
	"que":             QUE,
	"y":               Y,
	"o":               O,
	"no":              NO,
}

// displayNames holds the canonical PSeInt spelling of each keyword, used
// when a keyword has to be shown to the user (e.g. in diagnostics).
var displayNames = map[Keyword]string{
	ALGORITMO: "Algoritmo", FINALGORITMO: "FinAlgoritmo",
	PROCESO: "Proceso", FINPROCESO: "FinProceso",
	QUE: "Que", ESCRIBIR: "Escribir", IMPRIMIR: "Imprimir", LEER: "Leer", MOD: "MOD",
	SI: "Si", ENTONCES: "Entonces", SINO: "SiNo", FINSI: "FinSi",
	SEGUN: "Segun", FINSEGUN: "FinSegun", DE: "De", OTRO: "Otro", MODO: "Modo",
	DEFINIR: "Definir", COMO: "Como",
	ENTERO: "Entero", CADENA: "Cadena", TEXTO: "Texto", CARACTER: "Caracter",
	LOGICO: "Logico", REAL: "Real", NUMERO: "Numero", NUMERICO: "Numerico",
	DIMENSIONAR: "Dimensionar", DIMENSION: "Dimension", REDIMENSIONAR: "Redimensionar",
	MIENTRAS: "Mientras", HACER: "Hacer", FINMIENTRAS: "FinMientras",
	PARA: "Para", HASTA: "Hasta", CON: "Con", PASO: "Paso", FINPARA: "FinPara",
	REPETIR: "Repetir", SIN: "Sin", SALTAR: "Saltar",
	BORRAR: "Borrar", LIMPIAR: "Limpiar", PANTALLA: "Pantalla",
	ESPERAR: "Esperar", TECLA: "Tecla", SEGUNDOS: "Segundos", MILISEGUNDOS: "Milisegundos",
	FIN: "Fin", Y: "Y", O: "O", NO: "NO", CASO: "Caso",
	FUNCION: "Funcion", FINFUNCION: "FinFuncion", POR: "Por", REFERENCIA: "Referencia",
	SUBPROCESO: "SubProceso", FINSUBPROCESO: "FinSubProceso",
}

// Display returns the keyword as a PSeInt programmer would write it.
func (k Keyword) Display() string {
	if name, ok := displayNames[k]; ok {
		return name
	}
	return strings.ToLower(k.String())
}
