package neoparser

import (
	"pseint-compiled/internal/lexer"
	"pseint-compiled/internal/models"
)

type Context struct {
	Tokens      []lexer.Token
	Pos         uint32
	Diagnostics []ParseError
}

func (c *Context) EOF() bool {
	return c.Pos >= uint32(len(c.Tokens))
}

func (c *Context) Peek() *lexer.Token {
	if c.EOF() {
		return nil
	}

	return &c.Tokens[c.Pos]
}

// Peeks at a token a number of indices (offset) from the current position!
func (c *Context) PeekAt(offset uint32) *lexer.Token {
	pos := c.Pos + offset

	if pos >= uint32(len(c.Tokens)) {
		return nil
	}

	return &c.Tokens[pos]
}

func (c *Context) Advance() {
	if !c.EOF() {
		c.Pos++
	}
}

func (c *Context) Mark() uint32 {
	return c.Pos
}

// Gets two absolute positions in the token tape,
// Gets the tokens located in both positions and joins their spans
func (c *Context) Span(start uint32, end uint32) models.Span {
	/* Get span from other token */
	length := uint32(len(c.Tokens))
	if start >= length || end >= length {
		return models.Span{}
	}

	t1 := c.Tokens[start]
	t2 := c.Tokens[end]

	return models.JoinSpans(t1.Span, t2.Span)
}

/* Checkpoint logic */

type Checkpoint struct {
	Pos             uint32
	DiagnosticCount int
	Cut             uint64
}

func (c *Context) Checkpoint() Checkpoint {
	return Checkpoint{
		Pos:             c.Pos,
		DiagnosticCount: len(c.Diagnostics),
	}
}

func (c *Context) Restore(cp Checkpoint) {
	c.Pos = cp.Pos
	c.Diagnostics = c.Diagnostics[:cp.DiagnosticCount]
}
