package parser

import (
	"pascc/internal/ast"
	"pascc/internal/lexer"
)

func (p *Parser) parse_post(last_expr ast.Expr) ast.Expr {
	for {
		switch {
		case p.check(lexer.L_PARENTHESES):
			last_expr = p.parse_call(last_expr)

		case p.check(lexer.L_BRACKET):
			last_expr = p.parse_index(last_expr)

		default:
			return last_expr
		}

		if last_expr == nil {
			return nil
		}
	}
}

func (p *Parser) parse_call(last_expr ast.Expr) ast.Expr {
	start := p.mark()

	p.eat_token(lexer.L_PARENTHESES)

	args := []ast.Expr{}
	for !p.check(lexer.R_PARENTHESES) {
		arg, err := p.parse_expression()
		if err != nil {
			return nil
		}
		args = append(args, arg)

		if _, err := p.eat_token(lexer.COMMA); err != nil {
			if _, err := p.eat_token(lexer.R_PARENTHESES); err != nil {
				p.Report(p.currentSpan(), "Expected ')' or ','")
			}
			break
		}
	}

	return &ast.Call{
		NodeInfo:  p.infoFrom(start),
		Callable:  last_expr,
		Arguments: args,
	}
}

func (p *Parser) parse_index(last_expr ast.Expr) ast.Expr {
	start := p.mark()

	indices := []ast.Expr{}
	ok := true

	for p.check(lexer.L_BRACKET) {
		p.eat_token(lexer.L_BRACKET) // [

		index, err := p.parse_expression()
		if err != nil {
			ok = false
		}

		indices = append(indices, index)

		if _, err := p.eat_token(lexer.R_BRACKET); err != nil {
			p.Report(p.currentSpan(), "Expected ']'")
		}
	}

	if !ok {
		return nil
	}

	return &ast.Index{
		NodeInfo: p.infoFrom(start),
		Target:   last_expr,
		Indexes:  indices,
	}
}
