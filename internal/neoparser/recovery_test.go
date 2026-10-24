package neoparser

import (
	"pascc/internal/lexer"
	"testing"
)

func parseContext(t *testing.T, source string) *Context {
	t.Helper()
	tokens, err := lexer.NewLexer([]rune(source)).Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	return &Context{Tokens: tokens}
}

func TestMainBlocks(t *testing.T) {
	for _, source := range []string{
		"algoritmo demo\nfinalgoritmo",
		"algoritmo demo\nx = f()\nfinalgoritmo",
		"algoritmo demo\n\nescribir 1\n\nleer x\nfinalgoritmo",
		"algoritmo demo\nsi verdadero entonces\nescribir 1\nsino\nleer x\nfinsi\nfinalgoritmo",
		"algoritmo demo\nmientras verdadero hacer\nfinmientras\npara i = 1 hasta 5 con paso 1 hacer\nfinpara\nfinalgoritmo",
		"algoritmo demo\nsegun x hacer\n1:\nescribir 1\n2:\nleer x\nde otro modo:\nescribir 3\nfinsegun\nfinalgoritmo",
	} {
		t.Run(source, func(t *testing.T) {
			ctx := parseContext(t, source)
			result := MainFunction().Match(ctx)
			if !result.OK() || len(ctx.Diagnostics) != 0 || !ctx.EOF() {
				t.Fatalf("match=%+v diagnostics=%v pos=%d/%d", result, ctx.Diagnostics, ctx.Pos, len(ctx.Tokens))
			}
		})
	}
}

func TestRecoveryKeepsFollowingStatements(t *testing.T) {
	ctx := parseContext(t, "algoritmo demo\ndefinir como entero\nescribir 1 +\nleer x\nsi verdadero\nescribir\nsino\nescribir 2\nfinsi\nescribir 3\nfinalgoritmo")
	result := MainFunction().Match(ctx)
	if !result.OK() || !ctx.EOF() || len(ctx.Diagnostics) != 3 {
		t.Fatalf("match=%+v diagnostics=%v pos=%d", result, ctx.Diagnostics, ctx.Pos)
	}
	if len(result.Value.Stmts) != 3 {
		t.Fatalf("lost following statements: %v", result.Value.Stmts)
	}
}

func TestRequiredPartsFail(t *testing.T) {
	for _, source := range []string{"definir", "definir a", "definir a como", "leer", "escribir sin", "esperar 1", "borrar", "dimensionar a[", "si", "mientras", "para", "segun", "x =", "x = -", "x = f(1,)", "x = a[1", "x = (1", "x = 1."} {
		t.Run(source, func(t *testing.T) {
			ctx := parseContext(t, source)
			result := Statement().Match(ctx)
			if result.Kind != Failed || result.Err == nil {
				t.Fatalf("expected committed error, got %+v", result)
			}
		})
	}
}

func TestExpressionNoMatch(t *testing.T) {
	ctx := parseContext(t, "\n")
	result := Expression().Match(ctx)
	if result.Kind != NoMatch || ctx.Pos != 0 {
		t.Fatalf("got %+v at %d", result, ctx.Pos)
	}
}

func TestMissingInnerTerminatorPreservesOuter(t *testing.T) {
	ctx := parseContext(t, "algoritmo demo\nsi verdadero\nescribir 1\nfinalgoritmo")
	result := MainFunction().Match(ctx)
	if !result.OK() || !ctx.EOF() || len(ctx.Diagnostics) != 1 {
		t.Fatalf("got %+v diagnostics=%v pos=%d", result, ctx.Diagnostics, ctx.Pos)
	}
}

func TestSwitchHeaderRecovery(t *testing.T) {
	ctx := parseContext(t, "algoritmo demo\nsegun x\n1\n2:\nescribir 2\nfinsegun\nescribir 3\nfinalgoritmo")
	result := MainFunction().Match(ctx)
	if !result.OK() || !ctx.EOF() || len(ctx.Diagnostics) != 1 || len(result.Value.Stmts) != 2 {
		t.Fatalf("got %+v diagnostics=%v pos=%d", result, ctx.Diagnostics, ctx.Pos)
	}
}

func TestSpanAtEOF(t *testing.T) {
	ctx := parseContext(t, "123")
	result := Expression().Match(ctx)
	if !result.OK() || result.Value.NodeSpan() != ctx.Tokens[0].Span {
		t.Fatalf("lost span at EOF: %+v", result)
	}
}

func TestErrorLocationSurvivesRollbackAndRecovery(t *testing.T) {
	for _, withFallback := range []bool{false, true} {
		ctx := parseContext(t, "foo bar baz\nnext")
		pattern := PatternFunc[string](func(ctx *Context) Match[string] {
			checkpoint := ctx.Checkpoint()
			ctx.Advance()
			err := failureAt(ctx, ctx.Pos, "expected something else")
			ctx.Restore(checkpoint)
			return Match[string]{Kind: Failed, Start: checkpoint.Pos, End: err.TokenPosition(), Err: err}
		})
		var recovered Match[string]
		if withFallback {
			recovered = RecoverWith(pattern, UntilNewline{}, func(err ParseError) string {
				if err.TokenPosition() != 1 {
					t.Fatalf("fallback lost position: %d", err.TokenPosition())
				}
				return "fallback"
			}).Match(ctx)
		} else {
			recovered = Recover(pattern, UntilNewline{}).Match(ctx)
		}
		if ctx.Pos != 3 || recovered.End != 3 || len(ctx.Diagnostics) != 1 {
			t.Fatalf("recovery failed: %+v at %d", recovered, ctx.Pos)
		}
		if ctx.Diagnostics[0].TokenPosition() != 1 {
			t.Fatalf("error moved during recovery: %d", ctx.Diagnostics[0].TokenPosition())
		}
	}
}
