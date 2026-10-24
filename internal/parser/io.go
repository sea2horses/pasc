package parser

import (
	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/lexer"
)

func (p *Parser) parse_write() (ast.Stmt, error) {
	diagnostics.Dbg("Parsing write...")
	start := p.mark()
	_, err := p.eat_keyword(lexer.ESCRIBIR)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.ESCRIBIR})
	}

	newline := true
	if p.atKeyword(lexer.SIN) {
		p.eat_keyword(lexer.SIN)

		_, err := p.eat_keyword(lexer.SALTAR)
		if err != nil {
			p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.SALTAR})
		}

		newline = false
	}
	/* Expressions to be printed */
	print, err := p.parse_expression_list()
	return &ast.Write{Print: print, NodeInfo: p.infoFrom(start), Newline: newline}, nil
}

func (p *Parser) parse_read() (ast.Stmt, error) {
	diagnostics.Dbg("Parsing read...")
	start := p.mark()
	_, err := p.eat_keyword(lexer.LEER)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.LEER})
	}
	/* Expressions to be read */
	into, err := p.parse_expression_list()
	return &ast.Read{Into: into, NodeInfo: p.infoFrom(start)}, nil
}

func (p *Parser) parse_clear_screen() (ast.Stmt, error) {
	start := p.mark()
	_, err := p.eat_keyword(lexer.BORRAR)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.BORRAR})
		return nil, err
	}

	_, err = p.eat_keyword(lexer.PANTALLA)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.PANTALLA})
		return nil, err
	}

	return &ast.ClearScreen{NodeInfo: p.infoFrom(start)}, nil
}
