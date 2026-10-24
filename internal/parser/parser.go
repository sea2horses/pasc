package parser

import (
	"pascc/internal/ast"
	"pascc/internal/diagnostics"
	"pascc/internal/lexer"
	"pascc/internal/models"
	"slices"
)

type Parser struct {
	Tokens   []lexer.Token
	position uint32
	diagnostics.DiagnosticEmitter
}

func (p *Parser) Parse() (ast.Node, []diagnostics.Diagnostic) {
	p.position = 0
	p.DiagnosticEmitter = diagnostics.NewDiagnosticEmitter()

	program, _ := p.parse_main_function()

	return program, slices.Clone(p.Diagnostics())
}

func (p *Parser) snapshot() Parser {
	clone := *p
	return clone
}

func (p *Parser) returnToSnapshot(snapshot Parser) {
	*p = snapshot
}

func (p *Parser) eof() bool {
	return p.position >= uint32(len(p.Tokens))
}

/* Returns whether if parser is at End of Line, if at EOF, returns false */
func (p *Parser) eol() bool {
	if p.eof() {
		return false
	}

	tok, _ := p.get()
	return tok.Type == lexer.NEWLINE
}

func (p *Parser) check(tok lexer.TokenType) bool {
	if p.eof() {
		return false
	}

	ntok, _ := p.get()
	return tok == ntok.Type
}

func (p *Parser) currentSpan() models.Span {
	if token, err := p.get(); err == nil {
		return token.Span
	}

	if len(p.Tokens) == 0 {
		return models.Span{}
	}

	last := p.Tokens[len(p.Tokens)-1].Span

	// Empty span immediately after the final token.
	return models.Span{
		Start: last.End,
		End:   last.End,
	}
}

func (p *Parser) at(position uint32) (*lexer.Token, error) {
	if position >= uint32(len(p.Tokens)) {
		return nil, ErrOutOfBounds
	}

	return &p.Tokens[position], nil
}

func (p *Parser) mark() uint32 {
	return p.position
}

func (p *Parser) PositionSnapshot() (uint32, diagnostics.DiagnosticInfo) {
	return p.mark(), p.Checkpoint()
}

func (p *Parser) spanFrom(start uint32) models.Span {
	// No tokens consumed.
	if start >= p.position {
		return models.Span{}
	}

	if start >= uint32(len(p.Tokens)) {
		return models.Span{}
	}

	endIndex := p.position - 1
	if endIndex >= uint32(len(p.Tokens)) {
		endIndex = uint32(len(p.Tokens)) - 1
	}

	first := p.Tokens[start].Span
	last := p.Tokens[endIndex].Span

	return models.Span{
		Start: first.Start,
		End:   last.End,
	}
}

func (p *Parser) infoFrom(start uint32) ast.NodeInfo {
	return ast.NodeInfo{
		Span: p.spanFrom(start),
	}
}

/* TODO: When golang 1.27 is out, change this to a generic method */
func parserTry[T any](p *Parser, fn func() (T, error)) (T, bool) {
	init_position := p.position
	val, err := fn()
	if err != nil {
		p.position = init_position
	}
	return val, err == nil
}

func (p *Parser) get() (*lexer.Token, error) {
	return p.at(p.position)
}

/* Runs a predicate against every token, if predicate was successful, returns true, if it got to EOL (end-of-line), returns false */
func (p *Parser) until(predicate func(lexer.Token) bool) bool {
	token, err := p.get()

	for err == nil {
		if predicate(*token) || p.eol() {
			break
		}

		p.position++
		token, err = p.get()
	}

	return !p.eof() && predicate(*token)
}

func (p *Parser) until_eof(predicate func(lexer.Token) bool) bool {
	token, err := p.get()

	for err == nil {
		if predicate(*token) {
			break
		}

		p.position++
		token, err = p.get()
	}

	return !p.eof() && predicate(*token)
}

/* Skips to token, return true if token type was found */
func (p *Parser) skip_to_token(token_type lexer.TokenType) bool {
	diagnostics.Dbg("Skipping to token: ", token_type)
	return p.until(
		func(token lexer.Token) bool { return token.Type == token_type },
	)
}

func (p *Parser) skip_to_keyword(keyword lexer.Keyword) bool {
	diagnostics.Dbg("Skipping to keyword: ", keyword)
	return p.until(
		func(token lexer.Token) bool {
			kw, ok := lexer.MapToKeyword([]rune(token.Value))
			return token.Type == lexer.KEYWORD && ok && kw == keyword
		},
	)
}

func (p *Parser) skip_to_nextline() bool {
	diagnostics.Dbg("Skipping to next line... Position: ", p.position)
	ok := p.skip_to_token(lexer.NEWLINE)
	p.eat_token(lexer.NEWLINE)
	diagnostics.Dbg("Ok status:", ok)
	diagnostics.Dbg("New position: ", p.position)
	return ok
}

func (p *Parser) skip_newlines() bool {
	diagnostics.Dbg("Skipping newlines...")
	return p.until_eof(
		func(token lexer.Token) bool {
			return token.Type != lexer.NEWLINE
		},
	)
}

func (p *Parser) eat_token(token_type lexer.TokenType) (string, error) {
	diagnostics.Dbg("Ate token: ", token_type)
	if p.eof() {
		return "", ErrUnexpectedEOF
	}

	token, _ := p.get()

	if token.Type != token_type {
		diagnostics.Dbg("Didn't go well!")
		return "", ErrExpectedToken{Got: token.Type, Expected: token_type}
	}

	p.position++

	return token.Value, nil
}

func (p *Parser) eat_keyword(expected lexer.Keyword) (*lexer.Keyword, error) {
	diagnostics.Dbg("Ate keyword: ", expected)
	if p.eof() {
		// p.Report(p.currentSpan(), "%s", ErrUnexpectedEOF)
		return nil, ErrUnexpectedEOF
	}

	tok, _ := p.get()

	keyword, ok := lexer.MapToKeyword([]rune(tok.Value))
	if !ok {
		return nil, ErrUnrecognizedKeyword
	}

	if keyword != expected {
		return nil, ErrExpectedKeyword{Expected: expected}
	}

	p.position++

	return &keyword, nil
}
