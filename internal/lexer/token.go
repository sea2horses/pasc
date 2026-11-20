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
	SEMICOLON     // ;

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
	UPTICK    // ^

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
	';':  SEMICOLON,
	'%':  MODULO,
	'^':  UPTICK,
	'\n': NEWLINE,
}

// Symbol returns the source character of a punctuation/operator token type.
func (t TokenType) Symbol() (string, bool) {
	if t == NEWLINE {
		return "", false
	}
	for r, kind := range charMap {
		if kind == t {
			return string(r), true
		}
	}
	return "", false
}
