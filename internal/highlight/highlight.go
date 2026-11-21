// Package highlight splits PSeInt source code into classified segments for
// syntax highlighting, and renders them for terminals (ANSI) and the web
// (HTML).
//
// Unlike the compiler, the highlighter never rejects input: code with
// syntax errors is highlighted as far as it can be, and characters the
// lexer cannot read are marked Invalid instead of being dropped. Joining
// the text of the segments always gives back the source exactly.
package highlight

import (
	"strings"
	"unicode"

	"pascc/internal/analyzer"
	"pascc/internal/lexer"
	"pascc/internal/semantic"
)

// Class is the syntactic role of a segment.
type Class uint8

const (
	Plain       Class = iota /* Whitespace between tokens */
	Keyword                  /* Si, Mientras, Definir, Escribir... */
	Operator                 /* <-, +, <=, and the word operators Y, O, NO, MOD */
	Punctuation              /* , ; : ( ) [ ] */
	Identifier               /* Variables and other names */
	Function                 /* Functions and subprocesses defined in the file */
	Builtin                  /* Predefined functions and constants: Raiz, Azar, PI... */
	Type                     /* Type names after Como: Entero, Real... */
	String
	Number
	Boolean
	Comment
	Invalid /* Characters the lexer could not read, such as an unterminated string */
)

// Segment is a run of source text with a single class.
type Segment struct {
	Text  string
	Class Class
}

// wordOperators are keywords that act as operators in expressions.
var wordOperators = map[lexer.Keyword]bool{
	lexer.Y: true, lexer.O: true, lexer.NO: true, lexer.MOD: true,
}

// Segments classifies all of src.
func Segments(src []rune) []Segment {
	lex := lexer.NewLexer(src)
	lex.KeepComments = true
	all, _ := lex.Tokenize()

	tokens := all[:0:0]
	for _, tok := range all {
		if tok.Type != lexer.NEWLINE {
			tokens = append(tokens, tok)
		}
	}
	functions := definedFunctions(tokens)

	var segments []Segment
	emit := func(text string, class Class) {
		if text == "" {
			return
		}
		if n := len(segments); n > 0 && segments[n-1].Class == class {
			segments[n-1].Text += text
			return
		}
		segments = append(segments, Segment{Text: text, Class: class})
	}

	pos := 0
	for i, tok := range tokens {
		start := int(tok.Span.Start.Offset)
		end := min(start+tokenLength(tok), len(src))
		if start < pos {
			continue
		}
		emitGap(src[pos:start], emit)
		emit(string(src[start:end]), classify(tokens, i, functions))
		pos = end
	}
	emitGap(src[pos:], emit)
	return segments
}

// emitGap writes the text between two tokens: whitespace is Plain, and
// anything else is something the lexer skipped because it could not read it.
func emitGap(gap []rune, emit func(string, Class)) {
	for len(gap) > 0 {
		space := unicode.IsSpace(gap[0])
		n := 1
		for n < len(gap) && unicode.IsSpace(gap[n]) == space {
			n++
		}
		class := Invalid
		if space {
			class = Plain
		}
		emit(string(gap[:n]), class)
		gap = gap[n:]
	}
}

func tokenLength(tok lexer.Token) int {
	n := len([]rune(tok.Value))
	if tok.Type == lexer.STRING_LITERAL {
		n += 2 /* The delimiters are not part of the value */
	}
	return n
}

func classify(tokens []lexer.Token, i int, functions map[string]bool) Class {
	tok := tokens[i]
	switch tok.Type {
	case lexer.KEYWORD:
		if kw, ok := lexer.MapToKeyword([]rune(tok.Value)); ok && wordOperators[kw] {
			return Operator
		}
		return Keyword
	case lexer.IDENTIFIER:
		name := semantic.NormalizeName(tok.Value)
		switch {
		case i > 0 && isKeyword(tokens[i-1], lexer.COMO) && semantic.IsBuiltinTypeName(name):
			return Type
		case functions[name]:
			return Function
		case analyzer.IsBuiltin(name):
			return Builtin
		}
		return Identifier
	case lexer.STRING_LITERAL:
		return String
	case lexer.NUMBER_LITERAL:
		return Number
	case lexer.BOOLEAN_LITERAL:
		return Boolean
	case lexer.COMMENT:
		return Comment
	case lexer.DOT:
		// The lexer splits 3.14 into 3 . 14
		if i+1 < len(tokens) && tokens[i+1].Type == lexer.NUMBER_LITERAL {
			return Number
		}
		return Operator
	case lexer.COMMA, lexer.SEMICOLON, lexer.COLON,
		lexer.L_PARENTHESES, lexer.R_PARENTHESES, lexer.L_BRACKET, lexer.R_BRACKET:
		return Punctuation
	}
	return Operator
}

// definedFunctions collects the names of the functions and subprocesses
// declared in the file. A header names the function either right before its
// parameter list ("Funcion r <- Sumar(a, b)") or, without parameters, as the
// last name on the line ("SubProceso Saludar").
func definedFunctions(tokens []lexer.Token) map[string]bool {
	functions := map[string]bool{}
	for i, tok := range tokens {
		if !isKeyword(tok, lexer.FUNCION) && !isKeyword(tok, lexer.SUBPROCESO) {
			continue
		}
		if i > 0 && isKeyword(tokens[i-1], lexer.FIN) {
			continue /* "Fin Funcion" */
		}
		line := tok.Span.Start.Line
		name := ""
		for j := i + 1; j < len(tokens) && tokens[j].Span.Start.Line == line; j++ {
			if tokens[j].Type == lexer.L_PARENTHESES {
				break
			}
			if tokens[j].Type == lexer.IDENTIFIER {
				name = tokens[j].Value
			}
		}
		if name != "" {
			functions[semantic.NormalizeName(name)] = true
		}
	}
	return functions
}

func isKeyword(tok lexer.Token, kw lexer.Keyword) bool {
	if tok.Type != lexer.KEYWORD {
		return false
	}
	got, ok := lexer.MapToKeyword([]rune(tok.Value))
	return ok && got == kw
}

// Text joins segments back into source text.
func Text(segments []Segment) string {
	var b strings.Builder
	for _, s := range segments {
		b.WriteString(s.Text)
	}
	return b.String()
}
