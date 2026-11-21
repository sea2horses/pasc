package formatter

import (
	"strings"
	"unicode"

	"pascc/internal/lexer"
	"pascc/internal/semantic"
)

// glyph is one printable unit of a line: a single token, or several
// punctuation tokens that together spell one operator ("<-", "<=", "**").
type glyph struct {
	text  string
	kind  lexer.TokenType
	kw    lexer.Keyword
	isKw  bool
	unary bool /* A prefix "-" or "!", which is written against its operand */
}

func (g *glyph) is(kw lexer.Keyword) bool { return g != nil && g.isKw && g.kw == kw }

// endsValue reports whether an operator after g is binary rather than unary.
func (g *glyph) endsValue() bool {
	switch g.kind {
	case lexer.IDENTIFIER, lexer.NUMBER_LITERAL, lexer.STRING_LITERAL, lexer.BOOLEAN_LITERAL,
		lexer.R_PARENTHESES, lexer.R_BRACKET:
		return true
	}
	return false
}

// compoundOperators are written as a single unit when their characters are
// adjacent in the source. Characters separated by whitespace are left
// separate, so "a < -1" is not rewritten as the visually different "a <- 1"
// (the parser treats both the same way, but a reader would not).
var compoundOperators = map[string]bool{
	"<-": true, "<=": true, ">=": true, "<>": true, "==": true,
	"!=": true, "&&": true, "||": true, "**": true, ":=": true,
}

func buildGlyphs(tokens []lexer.Token, src []rune) []glyph {
	glyphs := make([]glyph, 0, len(tokens))
	var prev *glyph /* last glyph that is not a comment */

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		g := glyph{kind: tok.Type, text: tokenText(tok, src)}

		switch tok.Type {
		case lexer.KEYWORD:
			g.kw, g.isKw = lexer.MapToKeyword([]rune(tok.Value))
			g.text = canonicalKeyword(tok.Value)
		case lexer.BOOLEAN_LITERAL:
			g.text = canonicalBool(tok.Value)
		case lexer.IDENTIFIER:
			if prev.is(lexer.COMO) && semantic.IsBuiltinTypeName(tok.Value) {
				g.text = titleCase(tok.Value)
			}
		case lexer.NUMBER_LITERAL, lexer.STRING_LITERAL, lexer.COMMENT:
		default:
			if i+1 < len(tokens) {
				next := tokens[i+1]
				pair := tok.Value + next.Value
				if compoundOperators[pair] && next.Span.Start.Offset == tok.Span.Start.Offset+1 {
					g.text = pair
					i++
				}
			}
			if (g.text == "-" || g.text == "!") && (prev == nil || !prev.endsValue()) {
				g.unary = true
			}
		}

		glyphs = append(glyphs, g)
		if g.kind != lexer.COMMENT {
			prev = &glyphs[len(glyphs)-1]
		}
	}
	capitalizePhrases(glyphs)
	return glyphs
}

// capitalizePhrases gives connectives their canonical casing where they are
// part of a fixed phrase ("De Otro Modo", "Por Referencia", "Hasta Que",
// "Mientras Que") rather than words of a colloquial expression.
func capitalizePhrases(glyphs []glyph) {
	for i := range glyphs {
		var prev, next *glyph
		if i > 0 {
			prev = &glyphs[i-1]
		}
		if i+1 < len(glyphs) {
			next = &glyphs[i+1]
		}
		g := &glyphs[i]
		if (g.is(lexer.DE) && next.is(lexer.OTRO)) ||
			(g.is(lexer.POR) && next.is(lexer.REFERENCIA)) ||
			(g.is(lexer.QUE) && (prev.is(lexer.HASTA) || prev.is(lexer.MIENTRAS))) {
			g.text = g.kw.Display()
		}
	}
}

func join(glyphs []glyph) string {
	var b strings.Builder
	for i := range glyphs {
		if i > 0 && spaceBetween(&glyphs[i-1], &glyphs[i]) {
			b.WriteByte(' ')
		}
		b.WriteString(glyphs[i].text)
	}
	return b.String()
}

func spaceBetween(prev, cur *glyph) bool {
	if prev.kind == lexer.COMMENT || cur.kind == lexer.COMMENT {
		return true
	}
	switch cur.text {
	case ",", ";", ":", ")", "]":
		return false
	}
	switch {
	case prev.text == "(" || prev.text == "[":
		return false
	case prev.unary:
		return false
	case cur.text == "(" || cur.text == "[":
		// Calls and indexing hug their target: f(x), v[i]. After a keyword or
		// an operator the bracket opens a grouping: Si (a > b), x <- (a + b).
		return prev.kind != lexer.IDENTIFIER && prev.text != ")" && prev.text != "]"
	case cur.kind == lexer.DOT && prev.kind == lexer.NUMBER_LITERAL, prev.kind == lexer.DOT:
		// The lexer splits 3.14 into 3 . 14
		return false
	}
	return true
}

// tokenText is the source text of a token, normalized for output.
func tokenText(tok lexer.Token, src []rune) string {
	switch tok.Type {
	case lexer.STRING_LITERAL:
		delimiter := string(src[tok.Span.Start.Offset])
		return delimiter + tok.Value + delimiter
	case lexer.COMMENT:
		return normalizeComment(tok.Value)
	}
	return tok.Value
}

// normalizeComment strips carriage returns and trailing whitespace from each
// line of a comment. The text is otherwise kept verbatim, including the
// indentation of the inner lines of a block comment.
func normalizeComment(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// keywordAliases are accepted spellings whose canonical casing is not the
// keyword's display name (which drops the accent or names the synonym).
var keywordAliases = map[string]string{
	"función":         "Función",
	"finfunción":      "FinFunción",
	"subalgoritmo":    "SubAlgoritmo",
	"finsubalgoritmo": "FinSubAlgoritmo",
}

// connectives keep the author's casing. They double as ordinary words in
// colloquial expressions ("x es mayor o igual que 10", "es divisible por"),
// where capitalizing them would read badly, and both casings of the
// operators ("y"/"Y", "mod"/"MOD") are common.
var connectives = map[lexer.Keyword]bool{
	lexer.Y: true, lexer.O: true, lexer.NO: true, lexer.MOD: true,
	lexer.QUE: true, lexer.POR: true, lexer.DE: true,
}

func canonicalKeyword(text string) string {
	lower := strings.ToLower(text)
	if alias, ok := keywordAliases[lower]; ok {
		return alias
	}
	kw, ok := lexer.MapToKeyword([]rune(text))
	if !ok || connectives[kw] {
		return text
	}
	if display := kw.Display(); strings.ToLower(display) == lower {
		return display
	}
	return text
}

// titleCase writes a type name the way PSeInt does: "ENTERO" -> "Entero".
func titleCase(text string) string {
	runes := []rune(strings.ToLower(text))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func canonicalBool(text string) string {
	switch strings.ToLower(text) {
	case "verdadero":
		return "Verdadero"
	case "falso":
		return "Falso"
	}
	return text
}
