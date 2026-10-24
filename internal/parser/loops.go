package parser

import (
	"errors"
	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/lexer"
)

func (p *Parser) parse_while() (ast.Stmt, error) {
	diagnostics.Dbg("Parsing while...")
	start, checkpoint := p.PositionSnapshot()
	_, err := p.eat_keyword(lexer.MIENTRAS)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.MIENTRAS})
	}

	condition, err := p.parse_expression()
	if err != nil {
		p.skip_to_keyword(lexer.HACER)
	}

	/* TODO: Make it a language option for optional 'Hacer' */
	_, err = p.eat_keyword(lexer.HACER)
	if err != nil {
		p.Warn(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.HACER})
	}

	_, err = p.eat_token(lexer.NEWLINE)
	if err != nil {
		p.Report(p.currentSpan(), "expected newline")
	}

	info := p.infoFrom(start)

	stmts, err := p.parse_statement_block(lexer.FINMIENTRAS)

	_, err = p.eat_keyword(lexer.FINMIENTRAS)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.FINMIENTRAS})
		p.Info(info.Span, "declared here")
	}

	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}

	return &ast.While{Condition: condition, Stmts: stmts, NodeInfo: info}, nil
}

func (p *Parser) parse_for() (ast.Stmt, error) {
	diagnostics.Dbg("Parsing for...")
	start, checkpoint := p.PositionSnapshot()

	_, err := p.eat_keyword(lexer.PARA)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.PARA})
	}

	/* we EXPECT an assignment */
	expr, err := p.parse_expression_statement()
	assignment := expr.(*ast.Assignment)

	if err != nil || assignment == nil {
		p.Report(p.currentSpan(), "Expected assignment")
		p.skip_to_nextline()
		return nil, errors.New("Expected assignment")
	}

	/* Type shit, let's go for the next one */
	_, err = p.eat_keyword(lexer.HASTA)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.HASTA})
	}

	/* Parse until expression */
	until, err := p.parse_expression()
	if err != nil {
		return nil, errors.New("Invalid expression")
	}

	var step ast.Expr = nil

	/* Is there a step?? */
	if p.atKeyword(lexer.CON) {
		_, err = p.eat_keyword(lexer.CON)
		if err != nil {
			p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.CON})
		}

		_, err = p.eat_keyword(lexer.PASO)
		if err != nil {
			p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.PASO})
		}

		step, err = p.parse_expression()
		if err != nil {
			return nil, errors.New("Invalid expression")
		}
	}

	_, err = p.eat_keyword(lexer.HACER)

	_, err = p.eat_token(lexer.NEWLINE)
	if err != nil {
		p.Report(p.currentSpan(), "expected newline")
		p.skip_to_nextline()
	}

	/* Get info from here */
	info := p.infoFrom(start)

	stmts, _ := p.parse_statement_block(lexer.FINPARA)

	_, err = p.eat_keyword(lexer.FINPARA)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{Expected: lexer.FINPARA})
	}

	if p.AnyErrorSince(checkpoint) {
		return nil, ErrInvalidStatement
	}

	return &ast.For{
		NodeInfo: info,
		Start:    assignment,
		Until:    until,
		Step:     step,
		Stmts:    stmts,
	}, nil
}
