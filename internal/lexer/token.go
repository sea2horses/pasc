package lexer

import "pascc/internal/models"

//go:generate stringer -type=TokenType
type TokenType int8

/* TODO: Add character tokens like , = etc. */
const (
	// Literals
	KEYWORD TokenType = iota
	IDENTIFIER
	STRING_LITERAL
	NUMBER_LITERAL
	BOOLEAN_LITERAL

	// Characters
	DOT           // .
	COMMA         // ,
	EQUALS        // =
	L_PARENTHESES // (
	R_PARENTHESES // )
	L_BRACKET     // [
	R_BRACKET     // ]
	COLON         // :

	// Binary Operators
	// Arithmetic
	PLUS     // +
	MINUS    // -
	ASTERISK // *
	SLASH    // /
	MODULO   // %
	// Comparison
	L_ANGLE   // <
	R_ANGLE   // >
	AMPERSAND // &
	PIPE      // |
	EX_MARK   // !

	// Special
	NEWLINE
	EOF
)

// Token struct
type Token struct {
	Type  TokenType   /* Type of the token */
	Value string      /* Value of the token */
	Span  models.Span /* Span of the token */
}

/* DON'T TOUCH! */
var NULL_TOKEN = Token{Type: EOF}

var charMap map[rune]TokenType = map[rune]TokenType{
	'.':  DOT,
	',':  COMMA,
	'=':  EQUALS,
	'(':  L_PARENTHESES,
	')':  R_PARENTHESES,
	'+':  PLUS,
	'-':  MINUS,
	'*':  ASTERISK,
	'/':  SLASH,
	'<':  L_ANGLE,
	'>':  R_ANGLE,
	'&':  AMPERSAND,
	'|':  PIPE,
	'!':  EX_MARK,
	'[':  L_BRACKET,
	']':  R_BRACKET,
	':':  COLON,
	'%':  MODULO,
	'\n': NEWLINE,
}
