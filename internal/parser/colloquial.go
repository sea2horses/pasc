package parser

import (
	"strings"

	"pascc/internal/ast"
	"pascc/internal/lexer"
	"pascc/internal/models"
	"pascc/internal/semantic"
)

// PSeInt's "lenguaje coloquial" lets conditions read like Spanish:
//
//	x ES PAR, x ES IMPAR, x ES POSITIVO, x ES NEGATIVO, x ES CERO
//	x ES y, x ES IGUAL A y, x ES DISTINTO DE y
//	x ES MAYOR QUE y, x ES MENOR O IGUAL QUE y, ...
//	x ES DIVISIBLE POR y, x ES MULTIPLO DE y
//
// These words are matched by their text instead of being reserved keywords,
// because names such as "par", "mayor" or "cero" are common variables. "es"
// can only appear here after a complete operand, where an identifier could
// never be valid, so recognizing it is unambiguous.

type colloquialKind uint8

const (
	colloquialNone colloquialKind = iota
	colloquialCompare
	colloquialEven
	colloquialOdd
	colloquialPositive
	colloquialNegative
	colloquialZero
	colloquialDivisible
)

// unary reports whether the form takes no right-hand operand.
func (k colloquialKind) unary() bool {
	switch k {
	case colloquialEven, colloquialOdd, colloquialPositive, colloquialNegative, colloquialZero:
		return true
	}
	return false
}

type colloquialForm struct {
	words []string
	kind  colloquialKind
	op    semantic.OperatorType
}

// Longer forms first, so "mayor o igual que" wins over "mayor que".
var colloquialForms = []colloquialForm{
	{[]string{"mayor", "o", "igual", "que"}, colloquialCompare, semantic.GREATER_EQ},
	{[]string{"menor", "o", "igual", "que"}, colloquialCompare, semantic.LESSER_EQ},
	{[]string{"mayor", "que"}, colloquialCompare, semantic.GREATER},
	{[]string{"menor", "que"}, colloquialCompare, semantic.LESSER},
	{[]string{"igual", "a"}, colloquialCompare, semantic.EQUALS},
	{[]string{"igual", "que"}, colloquialCompare, semantic.EQUALS},
	{[]string{"distinto", "de"}, colloquialCompare, semantic.NOT_EQUALS},
	{[]string{"distinta", "de"}, colloquialCompare, semantic.NOT_EQUALS},
	{[]string{"diferente", "de"}, colloquialCompare, semantic.NOT_EQUALS},
	{[]string{"divisible", "por"}, colloquialDivisible, semantic.EQUALS},
	{[]string{"multiplo", "de"}, colloquialDivisible, semantic.EQUALS},
	{[]string{"par"}, colloquialEven, semantic.EQUALS},
	{[]string{"impar"}, colloquialOdd, semantic.NOT_EQUALS},
	{[]string{"positivo"}, colloquialPositive, semantic.GREATER},
	{[]string{"positiva"}, colloquialPositive, semantic.GREATER},
	{[]string{"negativo"}, colloquialNegative, semantic.LESSER},
	{[]string{"negativa"}, colloquialNegative, semantic.LESSER},
	{[]string{"cero"}, colloquialZero, semantic.EQUALS},
	{nil, colloquialCompare, semantic.EQUALS}, // plain "x ES y"
}

var accentFolder = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u")

func foldWord(word string) string {
	return accentFolder.Replace(strings.ToLower(word))
}

// wordAt reports whether the token at offset is the given word, whether the
// lexer classified it as an identifier or as a keyword (de, que, por, o).
func wordAt(ctx *Context, offset uint32, word string) bool {
	tok := ctx.PeekAt(offset)
	if tok == nil || (tok.Type != lexer.IDENTIFIER && tok.Type != lexer.KEYWORD) {
		return false
	}
	return foldWord(tok.Value) == word
}

// matchColloquial consumes "es ..." and returns the operator it stands for.
func matchColloquial(ctx *Context) (*ast.Operator, colloquialKind, bool) {
	if !wordAt(ctx, 0, "es") {
		return nil, colloquialNone, false
	}
	start := ctx.Peek().Span
	for _, form := range colloquialForms {
		matched := true
		for i, word := range form.words {
			if !wordAt(ctx, uint32(i+1), word) {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		end := ctx.PeekAt(uint32(len(form.words))).Span
		for range len(form.words) + 1 {
			ctx.Advance()
		}
		op := &ast.Operator{NodeInfo: ast.NodeInfo{Span: models.JoinSpans(start, end)}, Type: form.op}
		return op, form.kind, true
	}
	return nil, colloquialNone, false
}

// buildColloquial turns a colloquial test into ordinary operations.
func buildColloquial(kind colloquialKind, op ast.Operator, lhs, rhs ast.Expr) ast.Expr {
	span := op.NodeSpan()
	number := func(value string) ast.Expr {
		return &ast.NumberLiteral{NodeInfo: ast.NodeInfo{Span: span}, Int: value}
	}
	binary := func(l ast.Expr, t semantic.OperatorType, r ast.Expr) ast.Expr {
		spans := []models.Span{l.NodeSpan(), r.NodeSpan(), span}
		return &ast.BinaryOperation{
			NodeInfo: ast.NodeInfo{Span: models.JoinSpans(spans...)},
			LHS:      l,
			Op:       ast.Operator{NodeInfo: ast.NodeInfo{Span: span}, Type: t},
			RHS:      r,
		}
	}
	switch kind {
	case colloquialEven, colloquialOdd:
		return binary(binary(lhs, semantic.MODULO, number("2")), op.Type, number("0"))
	case colloquialPositive, colloquialNegative, colloquialZero:
		return binary(lhs, op.Type, number("0"))
	case colloquialDivisible:
		return binary(binary(lhs, semantic.MODULO, rhs), semantic.EQUALS, number("0"))
	default:
		return binary(lhs, op.Type, rhs)
	}
}
