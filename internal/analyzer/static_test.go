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
