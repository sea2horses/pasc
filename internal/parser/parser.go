package parser

import (
	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/lexer"
	"pascc/internal/models"
	"slices"
)

// Parser exposes the same entry point as the original parser.
type Parser struct {
	Tokens []lexer.Token
	diagnostics.DiagnosticEmitter
}

// Parse parses one complete algorithm. It returns no AST when syntax errors
// were found, including errors recovered inside the algorithm's body.
func (p *Parser) Parse() (ast.Node, []diagnostics.Diagnostic) {
	p.DiagnosticEmitter = diagnostics.NewDiagnosticEmitter()
	ctx := &Context{Tokens: p.Tokens}
	skipNewlines(ctx)
	result := Expected(Program(), ExpectedKeyword{Keyword: lexer.ALGORITMO}).Match(ctx)
	if !result.OK() {
		ctx.Diagnostics = append(ctx.Diagnostics, ensureError(ctx, result, ExpectedKeyword{Keyword: lexer.ALGORITMO}))
	} else {
		skipNewlines(ctx)
		if !ctx.EOF() {
			ctx.Diagnostics = append(ctx.Diagnostics, failureAt(ctx, ctx.Pos, ExpectedEndOfProgram))
		}
	}
	for _, err := range ctx.Diagnostics {
		pos := err.TokenPosition()
		span := models.Span{}
		if pos < uint32(len(p.Tokens)) {
			span = p.Tokens[pos].Span
		} else if last := lastSignificantToken(p.Tokens); last != nil {
			// Point just past the last real token instead of at trailing blank
			// lines, which may not exist as visible columns (e.g. CRLF files).
			end := last.Span.End
			span = models.Span{Start: end, End: end}
		}
		p.Report(span, "%s", err.Error())
	}
	ds := slices.Clone(p.Diagnostics())
	if len(ds) > 0 || !result.OK() {
		return nil, ds
	}
	return result.Value, ds
}

func lastSignificantToken(tokens []lexer.Token) *lexer.Token {
	for i := len(tokens) - 1; i >= 0; i-- {
		if tokens[i].Type != lexer.NEWLINE {
			return &tokens[i]
		}
	}
	if len(tokens) > 0 {
		return &tokens[len(tokens)-1]
	}
	return nil
}
