package parser

import (
	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/lexer"
	"pascc/internal/semantic"
)

func (p *Parser) parse_operator_type() semantic.OperatorType {
	init_position := p.position
	tok, err := p.get()
	if err != nil {
		return semantic.UNRECOGNIZED
	}

	switch tok.Type {
	case lexer.PLUS:
		p.eat_token(lexer.PLUS)
		return semantic.ADD
	case lexer.MINUS:
		p.eat_token(lexer.MINUS)
		return semantic.SUBTRACT
	case lexer.ASTERISK:
		p.eat_token(lexer.ASTERISK)
		return semantic.MULTIPLY
	case lexer.SLASH:
		p.eat_token(lexer.SLASH)
		return semantic.DIVIDE
	case lexer.L_ANGLE:
		p.eat_token(lexer.L_ANGLE)

		if tok, err := p.get(); err == nil && tok.Type == lexer.EQUALS {
			p.eat_token(lexer.EQUALS)
			return semantic.LESSER_EQ
		}
		return semantic.LESSER
	case lexer.R_ANGLE:
		p.eat_token(lexer.R_ANGLE)
		if tok, err := p.get(); err == nil && tok.Type == lexer.EQUALS {
			p.eat_token(lexer.EQUALS)
			return semantic.GREATER_EQ
		}
		return semantic.GREATER
	case lexer.EQUALS:
		p.eat_token(lexer.EQUALS)
		if tok, err := p.get(); err == nil && tok.Type == lexer.EQUALS {
			p.eat_token(lexer.EQUALS)
			return semantic.EQUALS
		}
		/* Restart position */
		p.position = init_position
		return semantic.UNRECOGNIZED
	case lexer.MODULO:
		p.eat_token(lexer.MODULO)
		return semantic.MODULO
	case lexer.EX_MARK:
		p.eat_token(lexer.EX_MARK)
		if tok, err := p.get(); err == nil && tok.Type == lexer.EQUALS {
			p.eat_token(lexer.EQUALS)
			return semantic.NOT_EQUALS
		}
		return semantic.NOT
	/* This is fucking up everything */
	case lexer.KEYWORD:
		_, err := p.eat_keyword(lexer.MOD)
		if err != nil {
			/* Restart position */
			p.position = init_position
			return semantic.UNRECOGNIZED
		}
		return semantic.MODULO
	case lexer.AMPERSAND:
		p.eat_token(lexer.AMPERSAND)
		if tok, err := p.get(); err == nil && tok.Type == lexer.AMPERSAND {
			p.eat_token(lexer.AMPERSAND)
			return semantic.AND
		}
		/* Restart position */
		p.position = init_position
		return semantic.UNRECOGNIZED
	case lexer.PIPE:
		p.eat_token(lexer.PIPE)
		if tok, err := p.get(); err == nil && tok.Type == lexer.PIPE {
			p.eat_token(lexer.PIPE)
			return semantic.OR
		}
		/* Restart position */
		p.position = init_position
		return semantic.UNRECOGNIZED
	}
	return semantic.UNRECOGNIZED
}

func (p *Parser) parse_operator() *ast.Operator {
	diagnostics.Dbg("Parsing operator!")
	start := p.mark()
	op_type := p.parse_operator_type()
	diagnostics.Dbg("Resolved Type: ", op_type)

	if op_type == semantic.UNRECOGNIZED {
		return nil
	} else {
		return &ast.Operator{Type: op_type, NodeInfo: p.infoFrom(start)}
	}
}
