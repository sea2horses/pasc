package neoparser

import (
	"pseint-compiled/internal/ast"
	"pseint-compiled/internal/diagnostics"
	"pseint-compiled/internal/lexer"
	"pseint-compiled/internal/models"
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
	result := Expected(MainFunction(), ExpectedKeyword{Keyword: lexer.ALGORITMO}).Match(ctx)
	if !result.OK() {
		ctx.Diagnostics = append(ctx.Diagnostics, ensureError(ctx, result, "expected algorithm"))
	} else {
		skipNewlines(ctx)
		if !ctx.EOF() {
			ctx.Diagnostics = append(ctx.Diagnostics, failureAt(ctx, ctx.Pos, "expected end of input"))
		}
	}
	for _, err := range ctx.Diagnostics {
		pos := err.TokenPosition()
		span := models.Span{}
		if pos < uint32(len(p.Tokens)) {
			span = p.Tokens[pos].Span
		} else if len(p.Tokens) > 0 {
			end := p.Tokens[len(p.Tokens)-1].Span.End
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
