//go:build ignore

package main

import (
	"bufio"
	"iter"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

/* Errors */
const (
	ErrIncorrectType     = "No coinciden los tipos"
	ErrNonWholeInteger   = "No se puede declarar un decimal a un entero"
	ErrNoDimensions      = "No se puede hacer un arreglo sin dimensiones"
	ErrNegativeDimension = "Un arreglo no puede tener dimensiones negativas o iguales a 0"
	ErrWrongIndexCount   = "No se proveyeron suficientes indices"
	ErrIndexOutOfBounds  = "El indice esta fuera de rango"
	ErrCouldNotReadKey   = "No se pudo leer una tecla"
	ErrTermError         = "No se pudo preparar la terminal para leer una tecla"
	ErrRestoreTermError  = "No se pudo restaurar la terminal"
)

func RuntimeError(msg string) {
	panic(msg)
}

/* I/O */
var writer bufio.Writer = *bufio.NewWriter(os.Stdout)
var reader bufio.Reader = *bufio.NewReader(os.Stdin)

type Value interface {
	int32 | float64 | string | bool
}

func Timeout(duration time.Duration) {
	time.Sleep(duration)
}

func writeval(value any) {
	switch val := value.(type) {
	case int32:
		writer.WriteString(strconv.FormatInt(int64(val), 10))
	case string:
		writer.WriteString(val)
	case float64:
		s := strconv.FormatFloat(val, 'f', 10, 64)
		// Clean up if decimal
		if strings.Contains(s, ".") {
			s = strings.TrimRight(s, "0")
			s = strings.Trim(s, ".")
		}
		writer.WriteString(s)
	case bool:
		if val {
			writer.WriteString("VERDADERO")
		} else {
			writer.WriteString("FALSO")
		}
	}
}

func Write(values ...any) {
	for _, value := range values {
		writeval(value)
	}
	writer.WriteRune('\n')
}

func WriteContinuous(values ...any) {
	for _, value := range values {
		writeval(value)
	}
}

func Read[T Value](dest *T) {
	writer.WriteString("> ")
	writer.Flush()

	/* TODO: Catch error from ReadLine */
	input, _, _ := reader.ReadLine()

	switch d := any(dest).(type) {
	case *int32:
		result, err := strconv.ParseInt(string(input), 10, 32)
		if err != nil {
			RuntimeError(ErrIncorrectType)
		}
		*d = int32(result)
	case *string:
		*d = string(input)
	case *float64:
		result, err := strconv.ParseFloat(string(input), 64)
		if err != nil {
			RuntimeError(ErrIncorrectType)
		}
		*d = result
	case *bool:
		*d = func() bool {
			/* Pseint takes false when no input is given */
			if len(input) == 0 {
				return false
			}

			/* Try to match input to other strings */
			input := strings.ToLower(string(input))
			if input == "verdadero" {
				return true
			}

			if input == "falso" {
				return false
			}

			/* Finally, match to 0 or 1 */
			result, err := strconv.ParseInt(input, 10, 64)
			if err == nil {
				switch result {
				case 0:
					return false
				case 1:
					return true
				}
			}

			RuntimeError(ErrIncorrectType)
			/* UNREACHABLE! */
			return false
		}()
	}
}

/* Implicit Casting */

func ToIntegerExact(value float64) int32 {
	const maxExclusive = 9223372036854775808.0
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Trunc(value) != value ||
		value < float64(math.MinInt32) || value >= maxExclusive {
		RuntimeError(ErrNonWholeInteger)
	}
	return int32(value)
}

func IntegerToReal(value int32) float64 {
	return float64(value)
}

/* PSeInt built-ins */
const PI = math.Pi

func Sqrt(value float64) float64 { return math.Sqrt(value) }
func Abs(value float64) float64  { return math.Abs(value) }
func Ln(value float64) float64   { return math.Log(value) }
func Exp(value float64) float64  { return math.Exp(value) }
func Sin(value float64) float64  { return math.Sin(value) }
func Cos(value float64) float64  { return math.Cos(value) }
func Tan(value float64) float64  { return math.Tan(value) }
func ASin(value float64) float64 { return math.Asin(value) }
func ACos(value float64) float64 { return math.Acos(value) }
func ATan(value float64) float64 { return math.Atan(value) }
func Trunc(value float64) int32  { return ToIntegerExact(math.Trunc(value)) }
func Round(value float64) int32  { return ToIntegerExact(math.Round(value)) }

func Azar(limit int32) int32 {
	if limit <= 0 {
		RuntimeError("Azar requiere un límite mayor que cero")
	}
	return int32(rand.Int31n(limit))
}

func Aleatorio(minimum, maximum int32) int32 {
	if maximum < minimum {
		RuntimeError("Aleatorio requiere que el límite inferior no supere al superior")
	}
	span := int64(maximum) - int64(minimum) + 1
	return int32(int64(minimum) + rand.Int63n(span))
}

func Longitud(value string) int32    { return int32(len([]rune(value))) }
func Mayusculas(value string) string { return strings.ToUpper(value) }
func Minusculas(value string) string { return strings.ToLower(value) }

func Subcadena(value string, start, end int32) string {
	runes := []rune(value)
	start -= array_bias
	end -= array_bias
	if start < 0 || end < start || end >= int32(len(runes)) {
		RuntimeError("Subcadena: posición fuera de los límites de la cadena")
	}
	return string(runes[start : end+1])
}

func Concatenar(first, second string) string { return first + second }

func ConvertirANumero(value string) float64 {
	number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		RuntimeError(ErrIncorrectType)
	}
	return number
}

func ConvertirATexto(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }

func HoraActual() int32 {
	now := time.Now()
	return int32(now.Hour()*10000 + now.Minute()*100 + now.Second())
}

func FechaActual() int32 {
	now := time.Now()
	return int32(now.Year()*10000 + int(now.Month())*100 + now.Day())
}

/* Arrays */
const array_bias = 0

type Tensor[T Value] struct {
	shape   []int32
	strides []int32
	data    []T
}

func NewTensor[T Value](shape ...int32) *Tensor[T] {
	if len(shape) == 0 {
		RuntimeError(ErrNoDimensions)
	}

	var size int32 = 1
	for _, dim := range shape {
		if dim <= 0 {
			RuntimeError(ErrNegativeDimension)
		}

		size *= dim
	}

	strides := make([]int32, len(shape))
	var stride int32 = 1
	for i := len(shape) - 1; i >= 0; i-- {
		strides[i] = stride
		stride *= shape[i]
	}

	return &Tensor[T]{
		shape:   append([]int32(nil), shape...),
		strides: strides,
		data:    make([]T, size),
	}
}

func DimensionSize[T Value](t *Tensor[T], dim int32) int32 {
	return int32(t.shape[dim])
}

func IndexTensor[T Value](t *Tensor[T], indexes ...int32) *T {
	if len(indexes) != len(t.shape) {
		RuntimeError(ErrWrongIndexCount)
	}

	/* Index bias */
	var offset int32 = 0

	for i, index := range indexes {
		/* Turn into real indexes */
		index = index - array_bias

		if index < 0 || index >= int32(t.shape[i]) {
			RuntimeError(ErrIndexOutOfBounds)
		}

		offset += index * int32(t.strides[i])
	}

	return &t.data[offset]
}

func IterTensor[T Value](tensor *Tensor[T]) iter.Seq2[T, int32] {
	return func(yield func(T, int32) bool) {
		for i := 0; i < len(tensor.data); i++ {
			if !yield(tensor.data[i], int32(i+array_bias)) {
				return
			}
		}
	}
}

/* FOR LOOPS */
/* PseInt Has some weird behaviour with for-loops so let's do this! */

func MakeIterStepless(start float64, end float64) iter.Seq[float64] {
	var step float64
	if end >= start {
		step = 1.0
	} else {
		step = -1.0
	}

	return MakeIter(start, end, step)
}

func MakeIter(start float64, end float64, step float64) iter.Seq[float64] {
	return func(yield func(float64) bool) {
		for i := start; (step > 0 && i <= end) || (step < 0 && i >= end); i += step {
			if !yield(i) {
				return
			}
		}
	}
}

/* Extra */

func ClearScreen() {
	writer.Flush()
	writer.WriteString("\x1b[2J\x1b[H")
}
