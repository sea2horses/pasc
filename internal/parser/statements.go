package parser

import (
	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/lexer"
	"slices"
)

func (p *Parser) parse_statement_block(terminators ...lexer.Keyword) ([]ast.Stmt, error) {
	diagnostics.Dbg("Parsing statement block...")
	var stmts []ast.Stmt
	for {
		p.skip_newlines()
		if p.eof() {
			break
		}

		/* Manual termination */
		tk, _ := p.get()
		if kw, ok := lexer.MapToKeyword([]rune(tk.Value)); ok && slices.Contains(terminators, kw) {
			break
		}

		stmt, err := p.parse_statement()

		if stmt != nil {
			diagnostics.Dbg("Statement was successful")
			_, err = p.eat_token(lexer.NEWLINE)
			if err != nil {
				p.Report(p.currentSpan(), "expected newline after statement")
			}

			stmts = append(stmts, stmt)
		} else {
			p.Report(p.currentSpan(), "extraneous statement")
			p.skip_to_nextline()
		}
	}
	return stmts, nil
}

func (p *Parser) parse_statement() (ast.Stmt, error) {
	diagnostics.Dbg("Parsing statement...")
	tk, _ := p.get()
	diagnostics.Dbgfmt("On token: %+v\n", tk)
	token, err := p.get()
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case lexer.KEYWORD:
		return p.parse_keyword()
	case lexer.IDENTIFIER, lexer.STRING_LITERAL, lexer.NUMBER_LITERAL,
		lexer.BOOLEAN_LITERAL, lexer.L_PARENTHESES, lexer.PLUS, lexer.MINUS,
		lexer.EX_MARK:
		return p.parse_expression_statement()
	}

	return nil, nil
}

func (p *Parser) parse_main_function() (*ast.MainFunction, error) {
	diagnostics.Dbg("Parsing main functions...")
	start, checkpoint := p.PositionSnapshot()

	var err error
	var fn_name string

	/* Did the user properly initialize the body? */
	init := true

	/* Set the proper structure */
	_, err = p.eat_keyword(lexer.ALGORITMO)
	if err != nil {
		init = false
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.ALGORITMO})
	}

	fn_name, err = p.eat_token(lexer.IDENTIFIER)
	if err != nil && init {
		p.Report(p.currentSpan(), "%s", ErrExpectedToken{Expected: lexer.IDENTIFIER})
	}

	_, err = p.eat_token(lexer.NEWLINE)
	if err != nil {
		p.Report(p.currentSpan(), "expected newline")
	}

	info := p.infoFrom(start)

	stmts, err := p.parse_statement_block(lexer.FINALGORITMO)
	/* TODO: Remove this */
	if err != nil {
		return nil, err
	}

	_, err = p.eat_keyword(lexer.FINALGORITMO)
	if err != nil && init {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.FINALGORITMO})
		p.Info(info.Span, "declared here")
	}

	/* If any errors were emitted, fuck everything sideways */
	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}

	return &ast.MainFunction{Name: fn_name, Stmts: stmts, NodeInfo: info}, nil
}

/* Can either be a function call or an assignment */
func (p *Parser) parse_expression_statement() (ast.Stmt, error) {
	start, checkpoint := p.PositionSnapshot()

	/* LHS */
	target, err := p.parse_expression()
	if err != nil {
		return nil, err
	}

	/* Assignment */
	if tk, err := p.get(); err == nil && tk.Type == lexer.EQUALS {
		p.eat_token(lexer.EQUALS)
		content, err := p.parse_expression()
		if err != nil {
			return nil, err
		}

		return &ast.Assignment{Target: target, Content: content, NodeInfo: p.infoFrom(start)}, nil
	}

	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}

	return &ast.ExprStmt{Expr: target, NodeInfo: p.infoFrom(start)}, nil
}

func (p *Parser) parse_keyword() (ast.Stmt, error) {
	diagnostics.Dbg("Parsing keyword...")
	token, err := p.get()

	if err != nil {
		return nil, err
	}

	if token.Type != lexer.KEYWORD {
		return nil, ErrExpectedToken{Got: token.Type, Expected: lexer.KEYWORD}
	}

	kw, ok := lexer.MapToKeyword([]rune(token.Value))
	if !ok {
		return nil, ErrUnrecognizedKeyword
	}

	switch kw {
	case lexer.DEFINIR:
		return p.parse_declaration()
	case lexer.ESCRIBIR:
		return p.parse_write()
	case lexer.LEER:
		return p.parse_read()
	case lexer.SI:
		return p.parse_if()
	case lexer.MIENTRAS:
		return p.parse_while()
	case lexer.SEGUN:
		return p.parse_switch()
	case lexer.BORRAR:
		return p.parse_clear_screen()
	case lexer.DIMENSIONAR:
		return p.parse_dimension()
	case lexer.PARA:
		return p.parse_for()
	case lexer.ESPERAR:
		return p.parse_timeout()
	}

	return nil, nil
}

func (p *Parser) parse_declaration() (ast.Stmt, error) {
	start, checkpoint := p.PositionSnapshot()
	/* GO */
	_, err := p.eat_keyword(lexer.DEFINIR)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{lexer.DEFINIR})
	}

	names := []string{}

	for {
		/* Var name */
		name, err := p.eat_token(lexer.IDENTIFIER)
		if err != nil {
			p.Report(p.currentSpan(), "expected variable name")
		} else {
			names = append(names, name)
		}

		if !p.check(lexer.COMMA) {
			break
		}

		p.eat_token(lexer.COMMA)
	}

	_, err = p.eat_keyword(lexer.COMO)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{lexer.COMO})
	}

	typeref, err := p.parse_type()
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedType)
	}

	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}

	return &ast.Declaration{
		NodeInfo: p.infoFrom(start),
		Names:    names,
		Type:     typeref,
	}, nil
}

func (p *Parser) parse_dimension() (ast.Stmt, error) {
	start, checkpoint := p.PositionSnapshot()

	_, err := p.eat_keyword(lexer.DIMENSIONAR)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{lexer.DIMENSIONAR})
	}

	/* Array name */
	name, err := p.eat_token(lexer.IDENTIFIER)
	if err != nil {
		p.Report(p.currentSpan(), "expected array name")
	}

	/* Dimensions */
	_, err = p.eat_token(lexer.L_BRACKET)
	if err != nil {
		p.Report(p.currentSpan(), "expected '[' (to declare array dimensions)")
	}

	dims, err := p.parse_expression_list()

	_, err = p.eat_token(lexer.R_BRACKET)
	if err != nil {
		p.Report(p.currentSpan(), "expected ']' (to close array dimensions)")
	}

	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}

	return &ast.Dimension{
		NodeInfo:   p.infoFrom(start),
		Name:       name,
		Dimensions: dims,
	}, nil
}
