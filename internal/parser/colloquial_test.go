package parser

import (
	"testing"

	"pascc/internal/ast"
)

func parseExpression(t *testing.T, source string) ast.Expr {
	t.Helper()
	ctx := parseContext(t, source)
	result := Expression().Match(ctx)
	if !result.OK() {
		t.Fatalf("%q did not parse: %v", source, result.Err)
	}
	if !ctx.EOF() {
		t.Fatalf("%q left tokens unparsed at %d", source, ctx.Pos)
	}
	return result.Value
}

func TestColloquialExpressions(t *testing.T) {
	cases := map[string]string{
		"x es par":                 "((x % 2) == 0)",
		"x es impar":               "((x % 2) != 0)",
		"x Es Positivo":            "(x > 0)",
		"x es negativo":            "(x < 0)",
		"x es cero":                "(x == 0)",
		"x es 3":                   "(x == 3)",
		"x es igual a 3":           "(x == 3)",
		"x es distinto de 3":       "(x != 3)",
		"x es mayor que 3":         "(x > 3)",
		"x es menor que 3":         "(x < 3)",
		"x es mayor o igual que 3": "(x >= 3)",
		"x es menor o igual que 3": "(x <= 3)",
		"x es divisible por 3":     "((x % 3) == 0)",
		"x es múltiplo de 3":       "((x % 3) == 0)",
		"a + b es par":             "(((a + b) % 2) == 0)",
		"a es par y b es impar":    "(((a % 2) == 0) && ((b % 2) != 0))",
		"a es mayor que b + 1":     "(a > (b + 1))",
		"no a es par":              "!(((a % 2) == 0))",
	}
	for source, want := range cases {
		t.Run(source, func(t *testing.T) {
			if got := render(parseExpression(t, source)); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestColloquialWordsRemainUsableAsNames(t *testing.T) {
	// A variable named "es" in a space-separated list must not start a test.
	ctx := parseContext(t, "algoritmo demo\nes = 1\npar = es + 1\nescribir \"v \" es\nfinalgoritmo")
	p := Parser{Tokens: ctx.Tokens}
	if node, ds := p.Parse(); node == nil || len(ds) != 0 {
		t.Fatalf("diagnostics = %v", ds)
	}
}

// render prints an expression with explicit grouping and operator symbols.
func render(expr ast.Expr) string {
	switch node := expr.(type) {
	case *ast.BinaryOperation:
		return "(" + render(node.LHS) + " " + node.Op.Type.String() + " " + render(node.RHS) + ")"
	case *ast.UnaryOperation:
		return node.Op.Type.String() + "(" + render(node.Expr) + ")"
	case *ast.NumberLiteral:
		return node.Int
	case *ast.Variable:
		return node.Name
	}
	return expr.String()
}
