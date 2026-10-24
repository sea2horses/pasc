package parser

import (
	"pascc/internal/ast"
	"pascc/internal/lexer"
)

func (p *Parser) parse_type() (*ast.TypeRef, error) {
	start := p.mark()

	/* Check if it's a keyword or an identifier */
	tok, err := p.get()
	if err != nil {
		return nil, err
	}

	switch tok.Type {
	case lexer.IDENTIFIER:
		p.eat_token(lexer.IDENTIFIER)
		return &ast.TypeRef{NodeInfo: p.infoFrom(start), Name: tok.Value}, nil
	case lexer.KEYWORD:
		kw, _ := lexer.MapToKeyword([]rune(tok.Value))
		switch kw {
		case lexer.CADENA, lexer.ENTERO, lexer.LOGICO, lexer.REAL:
			p.eat_token(lexer.KEYWORD)
			return &ast.TypeRef{NodeInfo: p.infoFrom(start), Name: tok.Value}, nil
		}
	}

	p.Report(tok.Span, string(ErrExpectedType))
	return nil, ErrExpectedType
}
