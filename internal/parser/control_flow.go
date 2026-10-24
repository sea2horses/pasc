package parser

import (
	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/lexer"
)

func (p *Parser) parse_if() (ast.Stmt, error) {
	diagnostics.Dbg("Parsing if...")
	start, checkpoint := p.PositionSnapshot()
	_, err := p.eat_keyword(lexer.SI)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.SI})
	}

	condition, err := p.parse_expression()

	_, err = p.eat_keyword(lexer.ENTONCES)
	if err != nil {
		p.Warn(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.ENTONCES})
	}

	_, err = p.eat_token(lexer.NEWLINE)
	if err != nil {
		p.Report(p.currentSpan(), "expected newline")
		p.skip_to_nextline()
	}

	info := p.infoFrom(start)

	/* Parse if statement block */
	stmts, err := p.parse_statement_block(lexer.FINSI, lexer.SINO)

	var else_branch *ast.Else = nil

	/* Check for else branch */
	if tok, err := p.get(); err == nil && tok.Type == lexer.KEYWORD {
		if keyword, ok := lexer.MapToKeyword([]rune(tok.Value)); ok && keyword == lexer.SINO {
			else_branch, err = p.parse_else()
			if err != nil {
				return nil, err
			}
		}
	}

	_, err = p.eat_keyword(lexer.FINSI)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.FINSI})
		p.Info(info.Span, "%s", "declared here")
	}

	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}
	return &ast.If{Condition: condition, Stmts: stmts, Else: else_branch, NodeInfo: info}, nil
}

func (p *Parser) parse_else() (*ast.Else, error) {
	start, checkpoint := p.PositionSnapshot()

	_, err := p.eat_keyword(lexer.SINO)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.SINO})
	}

	_, err = p.eat_token(lexer.NEWLINE)
	if err != nil {
		p.Report(p.currentSpan(), "expected newline")
	}

	info := p.infoFrom(start)

	/* Parse if statement block */
	stmts, err := p.parse_statement_block(lexer.FINSI)

	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}
	return &ast.Else{Stmts: stmts, NodeInfo: info}, nil
}

func (p *Parser) parse_switch() (ast.Stmt, error) {
	diagnostics.Dbg("Parsing switch...")
	start, checkpoint := p.PositionSnapshot()

	_, err := p.eat_keyword(lexer.SEGUN)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.SEGUN})
	}

	expr, err := p.parse_expression()
	if err != nil {
		p.Report(p.currentSpan(), "expected expression")
	}

	if _, err = p.eat_keyword(lexer.HACER); err != nil {
		p.Warn(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.HACER})
	}
	if _, err = p.eat_token(lexer.NEWLINE); err != nil {
		p.Report(p.currentSpan(), "expected newline")
	}

	info := ast.NodeInfo{Span: p.spanFrom(start)}

	cases := []*ast.Case{}
	var defaultStmts []ast.Stmt

	for {
		p.skip_newlines()
		if p.eof() || p.atKeyword(lexer.FINSEGUN) {
			break
		}

		if p.atDefaultCase() {
			defaultStmts, err = p.parse_default_case()
			if err != nil {
				break
			}
			p.skip_newlines()
			break
		}

		var c *ast.Case
		c, err = p.parse_case()
		if err != nil {
			p.Report(p.currentSpan(), "expected case expression followed by ':'")
			p.skip_to_nextline()
			continue
		}
		cases = append(cases, c)
	}

	if _, err = p.eat_keyword(lexer.FINSEGUN); err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.FINSEGUN})
	}

	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}

	return &ast.Switch{
		NodeInfo: info,
		Base:     expr,
		Cases:    cases,
		Default:  defaultStmts,
	}, nil
}

func (p *Parser) try_parse_case() (*ast.Case, error) {
	start := p.mark()
	clause, err := p.parse_expression()
	if err != nil {
		return nil, err
	}
	if _, err = p.eat_token(lexer.COLON); err != nil {
		return nil, err
	}
	if _, err = p.eat_token(lexer.NEWLINE); err != nil {
		return nil, err
	}
	return &ast.Case{Clause: clause, NodeInfo: p.infoFrom(start)}, nil
}

func (p *Parser) parse_case() (*ast.Case, error) {
	diagnostics.Dbg("Parsing case...")
	start, checkpoint := p.PositionSnapshot()

	/* Parse the clause */
	_case, err := p.try_parse_case()
	if err != nil {
		return nil, err
	}

	info := p.infoFrom(start)

	/* Parse statements up to the next case label/default/end marker. */
	stmts, err := p.parse_switch_statement_block()

	if err != nil {
		return nil, err
	}

	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}

	return &ast.Case{Clause: _case.Clause, Stmts: stmts, NodeInfo: info}, nil
}

func (p *Parser) parse_default_case() ([]ast.Stmt, error) {
	if _, err := p.eat_keyword(lexer.DE); err != nil {
		return nil, err
	}
	if _, err := p.eat_keyword(lexer.OTRO); err != nil {
		return nil, err
	}
	if _, err := p.eat_keyword(lexer.MODO); err != nil {
		return nil, err
	}
	if _, err := p.eat_token(lexer.COLON); err != nil {
		return nil, err
	}
	if _, err := p.eat_token(lexer.NEWLINE); err != nil {
		return nil, err
	}
	return p.parse_switch_statement_block()
}

func (p *Parser) parse_switch_statement_block() ([]ast.Stmt, error) {
	var stmts []ast.Stmt
	for {
		p.skip_newlines()
		if p.eof() || p.atKeyword(lexer.FINSEGUN) || p.atDefaultCase() || p.lineHasTopLevelColon() {
			return stmts, nil
		}

		stmt, err := p.parse_statement()
		if err != nil || stmt == nil {
			p.Report(p.currentSpan(), "extraneous statement")
			p.skip_to_nextline()
			continue
		}
		stmts = append(stmts, stmt)
		if _, err = p.eat_token(lexer.NEWLINE); err != nil && !p.eof() {
			p.Report(p.currentSpan(), "expected newline after statement")
		}
	}
}

func (p *Parser) atKeyword(expected lexer.Keyword) bool {
	tok, err := p.get()
	if err != nil || tok.Type != lexer.KEYWORD {
		return false
	}
	kw, ok := lexer.MapToKeyword([]rune(tok.Value))
	return ok && kw == expected
}

func (p *Parser) atDefaultCase() bool {
	if !p.atKeyword(lexer.DE) || p.position+2 >= uint32(len(p.Tokens)) {
		return false
	}
	for offset, expected := range []lexer.Keyword{lexer.OTRO, lexer.MODO} {
		tok := p.Tokens[p.position+uint32(offset)+1]
		kw, ok := lexer.MapToKeyword([]rune(tok.Value))
		if tok.Type != lexer.KEYWORD || !ok || kw != expected {
			return false
		}
	}
	return true
}

// A top-level colon is the switch grammar's discriminator: `expr:` starts a
// case, while the same expression without a colon is an ordinary statement.
func (p *Parser) lineHasTopLevelColon() bool {
	depth := 0
	for i := p.position; i < uint32(len(p.Tokens)); i++ {
		switch p.Tokens[i].Type {
		case lexer.NEWLINE:
			return false
		case lexer.L_PARENTHESES, lexer.L_BRACKET:
			depth++
		case lexer.R_PARENTHESES, lexer.R_BRACKET:
			if depth > 0 {
				depth--
			}
		case lexer.COLON:
			return depth == 0
		}
	}
	return false
}
