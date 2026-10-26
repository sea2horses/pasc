package analyzer

import (
	"strings"
	"testing"

	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/generator"
	"pascc/internal/lexer"
	"pascc/internal/parser"
	"pascc/internal/semantic"
)

func TestAnalyzerAndGeneratorArrayFunctionRegressions(t *testing.T) {
	const source = `Funcion valor <- PrimerElemento(valores)
	valor = valores[0]
Fin Funcion

Algoritmo RegresionesCompilador
	Definir valor Como Entero
	Dimension valores(1), sinTipo(1)
	valores[0] = 7
	valor = PrimerElemento(valores)
	Escribir "valor=", valor, ","; Sin Saltar
	Escribir " fin"
	Escribir, sinTipo[0]
FinAlgoritmo
`

	lex := lexer.NewLexer([]rune(source))
	tokens, lexDiagnostics := lex.Tokenize()
	if lexDiagnostics != nil {
		t.Fatalf("lexer diagnostics: %v", lexDiagnostics)
	}
	parse := parser.Parser{Tokens: tokens}
	tree, parseDiagnostics := parse.Parse()
	if len(parseDiagnostics) != 0 {
		t.Fatalf("parser diagnostics: %v", parseDiagnostics)
	}
	typed, analyzerDiagnostics := New().Analyze(tree.(*ast.MainFunction))
	for _, diagnostic := range analyzerDiagnostics {
		if diagnostic.Level == diagnostics.ERROR {
			t.Fatalf("analyzer error: %v", diagnostic)
		}
	}

	generated, err := generator.NewGenerator().Gen(typed)
	if err != nil {
		t.Fatal(err)
	}
	code := string(generated)
	if strings.Count(code, "var valor int32") != 2 {
		t.Fatalf("return variable should be local to its function and independent in main:\n%s", code)
	}
	if !strings.Contains(code, "NewTensor[string](1)") {
		t.Fatalf("unconstrained dimension did not receive a concrete fallback:\n%s", code)
	}
}

func TestTypeUnificationRejectsNilWithoutPanicking(t *testing.T) {
	if unify(nil, semantic.Integer_T{}) {
		t.Fatal("unify accepted a nil target")
	}
	if unify(semantic.Integer_T{}, nil) {
		t.Fatal("unify accepted a nil source")
	}
	if _, ok := coerce_to(nil, semantic.Integer_T{}); ok {
		t.Fatal("coerce_to accepted a nil expression")
	}
}

func TestInvalidArrayIndexProducesDiagnosticWithoutPanicking(t *testing.T) {
	const source = `Algoritmo IndiceInvalido
Dimension valores(1)
Escribir valores[VERDADERO]
FinAlgoritmo
`
	lex := lexer.NewLexer([]rune(source))
	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	parse := parser.Parser{Tokens: tokens}
	tree, parseDiagnostics := parse.Parse()
	if len(parseDiagnostics) != 0 {
		t.Fatalf("parser diagnostics: %v", parseDiagnostics)
	}
	_, analyzerDiagnostics := New().Analyze(tree.(*ast.MainFunction))
	for _, diagnostic := range analyzerDiagnostics {
		if strings.Contains(diagnostic.Msg, "index must be an integer") {
			return
		}
	}
	t.Fatalf("expected invalid-index diagnostic, got %v", analyzerDiagnostics)
}
