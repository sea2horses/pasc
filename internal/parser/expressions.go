package parser

import (
	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/lexer"
)

func (p *Parser) parse_expression() (ast.Expr, error) {
	diagnostics.Dbg("Parsing expression...")
	tk, _ := p.get()
	diagnostics.Dbgfmt("On token: %+v\n", tk)
	postfix, err := p.read_infix_to_postfix()
	if err != nil {
		return nil, err
	}
	return p.postfix_to_expression(postfix)
}

func (p *Parser) parse_expression_list() ([]ast.Expr, error) {
	expr_list := []ast.Expr{}
	for {
		expr, _ := p.parse_expression()
		if expr != nil {
			expr_list = append(expr_list, expr)
		}

		if tok, err := p.get(); err != nil || tok.Type != lexer.COMMA {
			break
		}
		p.eat_token(lexer.COMMA)
	}
	return expr_list, nil
}
