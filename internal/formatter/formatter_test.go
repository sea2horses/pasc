package formatter

import (
	"errors"
	"strings"
	"testing"
)

func format(t *testing.T, src string) string {
	t.Helper()
	out, err := Format([]rune(src), Options{Indent: "    "})
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	again, err := Format([]rune(out), Options{Indent: "    "})
	if err != nil {
		t.Fatalf("second Format: %v", err)
	}
	if again != out {
		t.Fatalf("not idempotent:\n--- once\n%s--- twice\n%s", out, again)
	}
	return out
}

func expect(t *testing.T, src, want string) {
	t.Helper()
	if got := format(t, src); got != want {
		t.Fatalf("unexpected output\n--- got\n%s--- want\n%s", got, want)
	}
}

func TestFormat(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{
			name: "indentation, spacing and keyword casing",
			src: `algoritmo   demo
definir x,z como entero
x<-1+2*3
  si x>=7 y x<>8 entonces
escribir "grande",x
      sino
escribir 'chico'
finsi
finalgoritmo
`,
			want: `Algoritmo demo
    Definir x, z Como Entero
    x <- 1 + 2 * 3
    Si x >= 7 y x <> 8 Entonces
        Escribir "grande", x
    SiNo
        Escribir 'chico'
    FinSi
FinAlgoritmo
`,
		},
		{
			name: "comments are kept and indented",
			src: `// cabecera
Algoritmo demo
// dentro
x = 1   // al final
  /* bloque
     de varias lineas   */
Si x = 1 Entonces
//cuerpo
Escribir x
FinSi
FinAlgoritmo
`,
			want: `// cabecera
Algoritmo demo
    // dentro
    x = 1 // al final
    /* bloque
     de varias lineas   */
    Si x = 1 Entonces
        //cuerpo
        Escribir x
    FinSi
FinAlgoritmo
`,
		},
		{
			name: "blank lines collapse and trailing ones are dropped",
			src:  "\n\nAlgoritmo demo\n\n\n\nx = 1   \n\n\nz = 2\nFinAlgoritmo\n\n\n",
			want: "Algoritmo demo\n\n    x = 1\n\n    z = 2\nFinAlgoritmo\n",
		},
		{
			name: "unary operators, calls, indexing and decimals",
			src: `Algoritmo demo
Dimension v [ 3 ]
x = - 1
w = x * -2 - ( -x )
z = abs ( x ) + v [ 1 ] + 3.14
Si NO x = 1 Y !( w > 2 ) Entonces
Escribir -x , 2**3
FinSi
FinAlgoritmo
`,
			want: `Algoritmo demo
    Dimension v[3]
    x = -1
    w = x * -2 - (-x)
    z = abs(x) + v[1] + 3.14
    Si NO x = 1 Y !(w > 2) Entonces
        Escribir -x, 2 ** 3
    FinSi
FinAlgoritmo
`,
		},
		{
			name: "operators split in the source stay split",
			src: `Algoritmo demo
x<-1
Si x < -1 Entonces
x := 2
FinSi
FinAlgoritmo
`,
			want: `Algoritmo demo
    x <- 1
    Si x < -1 Entonces
        x := 2
    FinSi
FinAlgoritmo
`,
		},
		{
			name: "segun cases",
			src: `Algoritmo demo
Leer x
Segun x Hacer
caso 1 , 2 :
Escribir "a"
3: Escribir "b"
Escribir "c"
de otro modo:
Escribir "d"
FinSegun
FinAlgoritmo
`,
			want: `Algoritmo demo
    Leer x
    Segun x Hacer
        Caso 1, 2:
            Escribir "a"
        3: Escribir "b"
            Escribir "c"
        De Otro Modo:
            Escribir "d"
    FinSegun
FinAlgoritmo
`,
		},
		{
			name: "post-test loops and nested while",
			src: `Algoritmo demo
x = 0
Repetir
x = x + 1
Mientras Que x < 3 Hacer
x = x + 1
FinMientras
Hasta Que x > 10
Hacer
x = x - 1
Mientras Que x > 0
Para i <- 1 Hasta 10 Con Paso 2 Hacer
Escribir i
Fin Para
FinAlgoritmo
`,
			want: `Algoritmo demo
    x = 0
    Repetir
        x = x + 1
        Mientras Que x < 3 Hacer
            x = x + 1
        FinMientras
    Hasta Que x > 10
    Hacer
        x = x - 1
    Mientras Que x > 0
    Para i <- 1 Hasta 10 Con Paso 2 Hacer
        Escribir i
    Fin Para
FinAlgoritmo
`,
		},
		{
			name: "functions and two-word end keywords",
			src: `funcion r <- Sumar ( a , b por referencia )
r = a + b
fin funcion
subproceso Saludar()
Escribir "hola";   Sin Saltar
finsubproceso
Algoritmo demo
Saludar()
FinAlgoritmo
`,
			want: `Funcion r <- Sumar(a, b Por Referencia)
    r = a + b
Fin Funcion
SubProceso Saludar()
    Escribir "hola"; Sin Saltar
FinSubProceso
Algoritmo demo
    Saludar()
FinAlgoritmo
`,
		},
		{
			name: "one-line blocks keep their lines",
			src: `Algoritmo demo
Si x = 1 Entonces Escribir 1 FinSi
Si x = 2 Entonces
Escribir 2
SiNo Escribir 3
FinSi
FinAlgoritmo
`,
			want: `Algoritmo demo
    Si x = 1 Entonces Escribir 1 FinSi
    Si x = 2 Entonces
        Escribir 2
    SiNo Escribir 3
    FinSi
FinAlgoritmo
`,
		},
		{
			name: "colloquial expressions keep their connectives",
			src: `Algoritmo demo
Si n es mayor o igual que 10 y n es divisible por 2 Entonces
Escribir verdadero
FinSi
FinAlgoritmo
`,
			want: `Algoritmo demo
    Si n es mayor o igual que 10 y n es divisible por 2 Entonces
        Escribir Verdadero
    FinSi
FinAlgoritmo
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) { expect(t, tc.src, tc.want) })
	}
}

func TestFormatDefaultIndentIsTab(t *testing.T) {
	out, err := Format([]rune("Algoritmo a\nEscribir 1\nFinAlgoritmo\n"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if want := "Algoritmo a\n\tEscribir 1\nFinAlgoritmo\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestFormatKeepsCRLF(t *testing.T) {
	out := format(t, "Algoritmo a\r\nEscribir 1 // c\r\nFinAlgoritmo\r\n")
	if want := "Algoritmo a\r\n    Escribir 1 // c\r\nFinAlgoritmo\r\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestFormatRejectsInvalidPrograms(t *testing.T) {
	for _, src := range []string{
		"Algoritmo a\nEscribir \"sin cerrar\nFinAlgoritmo\n", // lexical error
		"Algoritmo a\nEscribir 1 $ 2\nFinAlgoritmo\n",        // unexpected char
		"Algoritmo a\nSi x Entonces\nFinAlgoritmo\n",         // missing FinSi
	} {
		_, err := Format([]rune(src), Options{})
		var syntax *SyntaxError
		if !errors.As(err, &syntax) || len(syntax.Diagnostics) == 0 {
			t.Errorf("Format(%q) = %v, want a SyntaxError", strings.SplitN(src, "\n", 2)[1], err)
		}
	}
}
