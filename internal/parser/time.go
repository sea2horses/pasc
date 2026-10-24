package parser

import (
	"errors"
	"pascc/internal/ast"
	"pascc/internal/lexer"
	"pascc/internal/semantic"
)

func (p *Parser) parse_timeout() (*ast.TimeOut, error) {
	start, checkpoint := p.PositionSnapshot()

	_, err := p.eat_keyword(lexer.ESPERAR)
	if err != nil {
		p.Report(p.currentSpan(), "%s", ErrExpectedKeyword{lexer.ESPERAR})
	}

	/* Time amount */
	time, err := p.parse_expression()
	if err != nil {
		p.Report(p.currentSpan(), "expected expression")
	}

	/* Different time units */
	unit, err := p.parse_time_units()
	if err != nil {
		p.Report(p.currentSpan(), "expected time unit")
	}

	if p.AnyErrorSince(checkpoint) {
		return nil, errors.New("invalid expression")
	}

	return &ast.TimeOut{
		NodeInfo: p.infoFrom(start),
		Amount:   time,
		Unit:     unit,
	}, nil
}

func (p *Parser) parse_time_units() (semantic.TimeUnit, error) {
	if p.atKeyword(lexer.SEGUNDOS) {
		p.eat_keyword(lexer.SEGUNDOS)
		return semantic.SEGUNDOS, nil
	} else if p.atKeyword(lexer.MILISEGUNDOS) {
		p.eat_keyword(lexer.MILISEGUNDOS)
		return semantic.MILISEGUNDOS, nil
	}

	return 0, errors.New("time unit not recognized")
}
