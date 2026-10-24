package neoparser

import (
	"pseint-compiled/internal/ast"
	"pseint-compiled/internal/diagnostics"
	"strings"
	"testing"
)

func TestParseExpressionStatements(t *testing.T) {
	ctx := parseContext(t, "\nalgoritmo demo\nfoo()\nfoo(1)[2]\nx = foo()\n1 + 2\nfinalgoritmo\n\n")
	p := Parser{Tokens: ctx.Tokens}
	node, ds := p.Parse()
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	main, ok := node.(*ast.MainFunction)
	if !ok || len(main.Stmts) != 4 {
		t.Fatalf("unexpected AST: %#v", node)
	}
	for _, i := range []int{0, 1, 3} {
		stmt, ok := main.Stmts[i].(*ast.ExprStmt)
		if !ok {
			t.Fatalf("statement %d is %T", i, main.Stmts[i])
		}
		if stmt.Span != stmt.Expr.NodeSpan() {
			t.Fatalf("incorrect expression statement span: %+v", stmt.Span)
		}
	}
	if _, ok := main.Stmts[2].(*ast.Assignment); !ok {
		t.Fatalf("assignment became %T", main.Stmts[2])
	}
}

func TestParseDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		source, message string
		count           int
		eof             bool
	}{
		{"", "expected keyword", 1, true},
		{"escribir 1", "expected keyword", 1, false},
		{"algoritmo", "expected token", 1, true},
		{"algoritmo demo\nfoo()\n", "FINALGORITMO", 1, true},
		{"algoritmo demo\nfinalgoritmo\nfoo()", "expected end of input", 1, false},
		{"algoritmo demo\nx =\nfoo()\nfinalgoritmo", "expected expression", 1, false},
		{"algoritmo demo\nfoo(1,)\nbar()\nfinalgoritmo", "expected item after separator", 1, false},
		{"algoritmo demo\nescribir\nleer\nfinalgoritmo", "expected expression", 2, false},
	} {
		t.Run(tc.source, func(t *testing.T) {
			ctx := parseContext(t, tc.source)
			p := Parser{Tokens: ctx.Tokens}
			node, ds := p.Parse()
			if node != nil || len(ds) != tc.count {
				t.Fatalf("node=%#v diagnostics=%v", node, ds)
			}
			if !strings.Contains(strings.ToLower(ds[0].Msg), strings.ToLower(tc.message)) {
				t.Fatal(ds[0])
			}
			for _, d := range ds {
				if d.Level != diagnostics.ERROR {
					t.Fatalf("wrong severity: %v", d)
				}
			}
			if tc.eof && len(ctx.Tokens) > 0 {
				end := ctx.Tokens[len(ctx.Tokens)-1].Span.End
				if ds[0].Span.Start != end || ds[0].Span.End != end {
					t.Fatalf("wrong EOF span: %+v", ds[0].Span)
				}
			}
			if !tc.eof && ds[0].Span.Start.Line == 0 {
				t.Fatalf("missing token location: %+v", ds[0])
			}
			if strings.Contains(tc.source, "foo(1,)") {
				for _, tok := range ctx.Tokens {
					if tok.Value == ")" && tok.Span.Start.Line == 2 && ds[0].Span != tok.Span {
						t.Fatalf("separator diagnostic is misplaced: %+v", ds[0])
					}
				}
			}
			// Calls reset emitter state and return independent diagnostic slices.
			ds[0].Msg = "changed by caller"
			_, again := p.Parse()
			if len(again) != tc.count || again[0].Msg == ds[0].Msg || len(p.Diagnostics()) != tc.count {
				t.Fatalf("stale diagnostics: %v", again)
			}
			p.Tokens = parseContext(t, "algoritmo ok\nfinalgoritmo").Tokens
			valid, ds := p.Parse()
			if valid == nil || len(ds) != 0 || len(p.Diagnostics()) != 0 {
				t.Fatalf("parser did not reset: %v", ds)
			}
		})
	}
}

func TestRecoveredDiagnosticSpans(t *testing.T) {
	ctx := parseContext(t, "algoritmo demo\ndefinir como entero\nfoo(1,)\nescribir 42\nfinalgoritmo\n")
	p := Parser{Tokens: ctx.Tokens}
	node, ds := p.Parse()
	if node != nil || len(ds) != 2 {
		t.Fatalf("node=%#v diagnostics=%v", node, ds)
	}
	// Both errors occur well before the parser's final cursor. Each diagnostic
	// must retain its own offending token, rather than the recovery boundary.
	for i, target := range []struct {
		line  uint32
		value string
	}{{2, "como"}, {3, ")"}} {
		found := false
		for _, tok := range ctx.Tokens {
			if tok.Span.Start.Line == target.line && tok.Value == target.value {
				found = true
				if ds[i].Span != tok.Span {
					t.Fatalf("diagnostic %d has span %+v, want %+v", i, ds[i].Span, tok.Span)
				}
			}
		}
		if !found {
			t.Fatalf("missing target token: %+v", target)
		}
	}
}
