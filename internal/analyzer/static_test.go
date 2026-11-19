package analyzer

import (
	"strings"
	"testing"

	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
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
