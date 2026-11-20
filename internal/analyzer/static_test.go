package analyzer

import (
	"strings"
	"testing"

	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/generator"
	"pascc/internal/lexer"
	"pascc/internal/parser"
)

func analyzeSource(t *testing.T, source string, base IndexBase) (*tast.MainFunction, []diagnostics.Diagnostic) {
	t.Helper()
	tokens, lexDiagnostics := lexer.NewLexer([]rune(source)).Tokenize()
	if lexDiagnostics != nil {
		t.Fatalf("lexer diagnostics: %v", lexDiagnostics)
	}
	p := parser.Parser{Tokens: tokens}
	tree, parseDiagnostics := p.Parse()
	if len(parseDiagnostics) != 0 {
		t.Fatalf("parser diagnostics: %v", parseDiagnostics)
	}
	a := New()
	a.IndexBase = base
	return a.Analyze(tree.(*ast.MainFunction))
}

func diagnosticsAt(ds []diagnostics.Diagnostic, level diagnostics.DiagnosticLevel, contains string) int {
	count := 0
	for _, d := range ds {
		if d.Level == level && strings.Contains(d.Msg, contains) {
			count++
		}
	}
	return count
}

func TestIndexBaseDeduction(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   int
	}{
		{"no arrays default to zero", `Algoritmo A
	Escribir 1
FinAlgoritmo`, 0},
		{"literal zero index", `Algoritmo A
	Dimensionar v[3]
	v[0] = 1
FinAlgoritmo`, 0},
		{"literal index equal to size", `Algoritmo A
	Dimensionar v[3]
	v[1] = 1
	v[3] = 3
FinAlgoritmo`, 1},
		{"loop from one to symbolic size", `Algoritmo A
	Definir n, i Como Entero
	Leer n
	Dimensionar v[n]
	Para i <- 1 Hasta n Hacer
		v[i] = i
	FinPara
FinAlgoritmo`, 1},
		{"loop from zero to size minus one", `Algoritmo A
	Definir n, i Como Entero
	Leer n
	Dimensionar v[n]
	Para i <- 0 Hasta n - 1 Hacer
		v[i] = i
	FinPara
FinAlgoritmo`, 0},
		{"one-based loop shifted down", `Algoritmo A
	Definir n, i Como Entero
	Leer n
	Dimensionar v[n]
	Para i <- 1 Hasta n Hacer
		v[i - 1] = i
	FinPara
FinAlgoritmo`, 0},
		{"constant-folded size", `Algoritmo A
	Definir n, i Como Entero
	n = 4
	Dimensionar v[n]
	Para i <- 1 Hasta 4 Hacer
		v[i] = i
	FinPara
FinAlgoritmo`, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			main, _ := analyzeSource(t, tc.source, IndexBaseAuto)
			if main.IndexBase != tc.want {
				t.Fatalf("deduced base %d, want %d", main.IndexBase, tc.want)
			}
		})
	}
}

func TestIndexBaseFlagOverridesDeduction(t *testing.T) {
	source := `Algoritmo A
	Dimensionar v[3]
	v[0] = 1
FinAlgoritmo`
	main, ds := analyzeSource(t, source, IndexBaseOne)
	if main.IndexBase != 1 {
		t.Fatalf("explicit base ignored: got %d", main.IndexBase)
	}
	if diagnosticsAt(ds, diagnostics.WARNING, "assuming") != 0 {
		t.Fatalf("explicit base should not warn about deduction: %v", ds)
	}
}

func TestIndexBaseMixedEvidenceWarns(t *testing.T) {
	source := `Algoritmo A
	Dimensionar v[3]
	v[0] = 1
	v[1] = 1
	v[3] = 1
FinAlgoritmo`
	_, ds := analyzeSource(t, source, IndexBaseAuto)
	if diagnosticsAt(ds, diagnostics.WARNING, "--index-base") != 1 {
		t.Fatalf("expected one mixed-base warning, got %v", ds)
	}
}

func TestRealToIntegerConversionChecks(t *testing.T) {
	cases := []struct {
		name             string
		body             string
		errors, warnings int
	}{
		{"constant fraction is an error", "k = 7 / 2", 1, 0},
		{"constant whole quotient is fine", "k = 8 / 2", 0, 0},
		{"unknown real warns", "Leer x\n\tk = x", 0, 1},
		{"Redon is safe", "Leer x\n\tk = Redon(x)", 0, 0},
		{"Redondear alias is safe", "Leer x\n\tk = Redondear(x)", 0, 0},
		{"Trunc is safe", "Leer x\n\tk = Trunc(x * 2)", 0, 0},
		{"integral loop variable is safe", "Dimensionar v[3]\n\tPara i <- 0 Hasta 2 Hacer\n\t\tv[i] = 1\n\tFinPara", 0, 0},
		{"fractional loop step warns", "Dimensionar v[3]\n\tPara i <- 0 Hasta 2 Con Paso 0.5 Hacer\n\t\tv[i] = 1\n\tFinPara", 0, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := "Algoritmo A\n\tDefinir k Como Entero\n\tDefinir x Como Real\n\t" + tc.body + "\nFinAlgoritmo"
			_, ds := analyzeSource(t, source, IndexBaseZero)
			if got := diagnosticsAt(ds, diagnostics.ERROR, "Redon()"); got != tc.errors {
				t.Errorf("errors = %d, want %d: %v", got, tc.errors, ds)
			}
			if got := diagnosticsAt(ds, diagnostics.WARNING, "Redon()"); got != tc.warnings {
				t.Errorf("warnings = %d, want %d: %v", got, tc.warnings, ds)
			}
		})
	}
}

func TestCompileTimeBoundsChecks(t *testing.T) {
	cases := []struct {
		name             string
		base             IndexBase
		body             string
		errors, warnings int
	}{
		{"zero-based last index", IndexBaseZero, "v[4] = 1", 0, 0},
		{"zero-based size index", IndexBaseZero, "v[5] = 1", 1, 0},
		{"one-based zero index", IndexBaseOne, "v[0] = 1", 1, 0},
		{"one-based size index", IndexBaseOne, "v[5] = 1", 0, 0},
		{"negative constant", IndexBaseZero, "v[-1] = 1", 1, 0},
		{"folded expression", IndexBaseZero, "v[n] = 1", 1, 0},
		{"loop overruns", IndexBaseZero, "Para i <- 0 Hasta n Hacer\n\t\tv[i] = 1\n\tFinPara", 0, 1},
		{"loop fits", IndexBaseZero, "Para i <- 0 Hasta n - 1 Hacer\n\t\tv[i] = 1\n\tFinPara", 0, 0},
		{"shifted loop fits", IndexBaseZero, "Para i <- 1 Hasta n Hacer\n\t\tv[i - 1] = 1\n\tFinPara", 0, 0},
		{"empty loop is ignored", IndexBaseZero, "Para i <- 10 Hasta 20 Con Paso -1 Hacer\n\t\tv[i] = 1\n\tFinPara", 0, 0},
		{"reassigned size is not folded", IndexBaseZero, "n = 9\n\tv[n] = 1", 0, 0},
		{"unknown index is not checked", IndexBaseZero, "Leer k\n\tv[k] = 1", 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := "Algoritmo A\n\tDefinir n, k Como Entero\n\tn = 5\n\tDimensionar v[n]\n\t" + tc.body + "\nFinAlgoritmo"
			_, ds := analyzeSource(t, source, tc.base)
			if got := diagnosticsAt(ds, diagnostics.ERROR, "out of bounds"); got != tc.errors {
				t.Errorf("errors = %d, want %d: %v", got, tc.errors, ds)
			}
			if got := diagnosticsAt(ds, diagnostics.WARNING, "out of bounds"); got != tc.warnings {
				t.Errorf("warnings = %d, want %d: %v", got, tc.warnings, ds)
			}
		})
	}
}

func TestNonPositiveDimensionIsAnError(t *testing.T) {
	_, ds := analyzeSource(t, "Algoritmo A\n\tDimensionar v[0]\nFinAlgoritmo", IndexBaseAuto)
	if diagnosticsAt(ds, diagnostics.ERROR, "at least 1") != 1 {
		t.Fatalf("expected dimension error, got %v", ds)
	}
}

func TestImplicitIntegerWidensToReal(t *testing.T) {
	source := `Algoritmo W
	Definir x Como Real
	Definir k Como Entero
	acum = 0
	total = 0
	Leer x
	acum = acum + x
	total = total + acum
	Dimensionar z[3]
	z[0] = 0
	z[1] = x / 3
	k = 1
FinAlgoritmo`
	main, ds := analyzeSource(t, source, IndexBaseZero)
	if n := diagnosticsAt(ds, diagnostics.WARNING, "Redon()"); n != 0 {
		t.Fatalf("widened variables should not need conversions, got %d warnings: %v", n, ds)
	}
	code, err := generator.NewGenerator().Gen(main)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"var acum float64", "var total float64", "NewTensor[float64]", "var k int32"} {
		if !strings.Contains(string(code), want) {
			t.Errorf("generated code is missing %q:\n%s", want, code)
		}
	}
}

func TestDeclaredIntegerIsNotWidened(t *testing.T) {
	source := `Algoritmo W
	Definir k Como Entero
	Definir x Como Real
	Leer x
	k = x
FinAlgoritmo`
	_, ds := analyzeSource(t, source, IndexBaseZero)
	if diagnosticsAt(ds, diagnostics.WARNING, "Redon()") != 1 {
		t.Fatalf("declared Entero must keep its type and warn: %v", ds)
	}
}

func TestGuardedLoopAccessDoesNotWarn(t *testing.T) {
	source := `Algoritmo A
	Definir i Como Entero
	Dimensionar v[5]
	Para i <- 0 Hasta 4 Hacer
		Si i > 0 Entonces
			v[i - 1] = v[i]
		FinSi
	FinPara
FinAlgoritmo`
	_, ds := analyzeSource(t, source, IndexBaseAuto)
	if n := diagnosticsAt(ds, diagnostics.WARNING, "out of bounds") + diagnosticsAt(ds, diagnostics.WARNING, "--index-base"); n != 0 {
		t.Fatalf("guarded access should not warn: %v", ds)
	}
}

func TestShiftedLoopIndexIsNotOneBasedEvidence(t *testing.T) {
	source := `Algoritmo A
	Definir i Como Entero
	Dimensionar v[5]
	v[0] = 1
	Para i <- 0 Hasta 3 Hacer
		Escribir v[i + 1]
	FinPara
FinAlgoritmo`
	main, ds := analyzeSource(t, source, IndexBaseAuto)
	if main.IndexBase != 0 || diagnosticsAt(ds, diagnostics.WARNING, "--index-base") != 0 {
		t.Fatalf("v[i + 1] should not look 1-based: base=%d %v", main.IndexBase, ds)
	}
}

func TestLoopEvidenceUsesTheMatchingEnd(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   int
	}{
		// The upper end equal to 1 is not "starting at 1".
		{"zero-based loop ending at one", `Algoritmo A
	Definir j Como Entero
	Dimensionar v[2]
	Para j <- 0 Hasta 1 Hacer
		v[j] = j
	FinPara
FinAlgoritmo`, 0},
		// The lower end equal to size - 1 is not "running to size - 1".
		{"one-based loop over size two", `Algoritmo A
	Definir i Como Entero
	Dimensionar v[2]
	Para i <- 1 Hasta 2 Hacer
		v[i] = i
	FinPara
FinAlgoritmo`, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			main, ds := analyzeSource(t, tc.source, IndexBaseAuto)
			if main.IndexBase != tc.want {
				t.Fatalf("deduced base %d, want %d", main.IndexBase, tc.want)
			}
			if diagnosticsAt(ds, diagnostics.WARNING, "--index-base") != 0 {
				t.Fatalf("unexpected mixed-base warning: %v", ds)
			}
		})
	}
}

func TestRedimensionar(t *testing.T) {
	cases := []struct {
		name, body, wantError string
	}{
		{"bounds follow the new size", "Dimensionar v[3]\n\tRedimensionar v[5]\n\tv[4] = 1", ""},
		{"out of bounds after shrinking", "Dimensionar v[5]\n\tRedimensionar v[2]\n\tv[3] = 1", "out of bounds"},
		{"must already be an array", "Definir k Como Entero\n\tRedimensionar k[3]", "isn't an array"},
		{"must exist", "Redimensionar w[3]", "hasn't been dimensioned"},
		{"rank cannot change", "Dimensionar v[3]\n\tRedimensionar v[3, 3]", "from 1 to 2"},
		{"size after a conditional resize is unknown", "Dimensionar v[3]\n\tSi Verdadero Entonces\n\t\tRedimensionar v[9]\n\tFinSi\n\tv[8] = 1", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			main, ds := analyzeSource(t, "Algoritmo A\n\t"+tc.body+"\nFinAlgoritmo", IndexBaseZero)
			if tc.wantError == "" {
				if diagnosticsAt(ds, diagnostics.ERROR, "") != 0 {
					t.Fatalf("unexpected errors: %v", ds)
				}
				code, err := generator.NewGenerator().Gen(main)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(code), "ResizeTensor(v, ") {
					t.Fatalf("missing ResizeTensor call:\n%s", code)
				}
				return
			}
			if diagnosticsAt(ds, diagnostics.ERROR, tc.wantError) != 1 {
				t.Fatalf("expected error %q, got %v", tc.wantError, ds)
			}
		})
	}
}

func TestParaCada(t *testing.T) {
	source := `Algoritmo A
	Definir suma, cada Como Entero
	Dimensionar v[3]
	v[0] = 1
	Para Cada x De v Hacer
		suma = suma + x
		x = 0
	FinPara
	Para Cada y2 en v
		Escribir y2
	FinPara
	Para cada <- 1 Hasta 2 Hacer
		Escribir cada
	FinPara
FinAlgoritmo`
	main, ds := analyzeSource(t, source, IndexBaseAuto)
	if diagnosticsAt(ds, diagnostics.ERROR, "") != 0 {
		t.Fatalf("unexpected errors: %v", ds)
	}
	code, err := generator.NewGenerator().Gen(main)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"for x := range TensorElements(v)", "*x = 0", "for y2 := range TensorElements(v)"} {
		if !strings.Contains(string(code), want) {
			t.Errorf("generated code is missing %q:\n%s", want, code)
		}
	}
}

func TestParaCadaNeedsAnArray(t *testing.T) {
	_, ds := analyzeSource(t, "Algoritmo A\n\tDefinir n Como Entero\n\tPara Cada x De n Hacer\n\t\tEscribir x\n\tFinPara\nFinAlgoritmo", IndexBaseAuto)
	if diagnosticsAt(ds, diagnostics.ERROR, "needs an array") != 1 {
		t.Fatalf("expected a non-array error, got %v", ds)
	}
}
