// Package formatter rewrites PSeInt source code in a canonical layout.
//
// The formatter works on the token stream rather than on the AST, so it
// keeps everything the parser throws away: comments, the line structure,
// string delimiters and the spelling the author chose for each construct
// (for example "Fin Funcion" versus "FinFuncion", or "<-" versus "="). It
// only changes:
//
//   - indentation, which follows the block structure;
//   - spacing between tokens on a line;
//   - the casing of keywords and logical literals ("finsi" -> "FinSi");
//   - blank lines: runs of them collapse into one, and leading and trailing
//     blank lines are removed;
//   - trailing whitespace, which is removed.
//
// Because the parser does not depend on spacing within a line, none of
// these changes can alter the meaning of a program. Format also re-lexes
// its output and checks that it produced the same tokens before returning.
package formatter

import (
	"errors"
	"strings"

	"pascc/internal/diagnostics"
	"pascc/internal/lexer"
	"pascc/internal/parser"
)

// Options controls the layout produced by Format.
type Options struct {
	// Indent is written once per nesting level. Empty means a tab.
	Indent string
}

// SyntaxError is returned when the source has lexical or syntax errors. The
// formatter refuses to touch such files: the lexer drops characters it does
// not understand, so formatting them could silently delete code.
type SyntaxError struct {
	Diagnostics []diagnostics.Diagnostic
}

func (e *SyntaxError) Error() string {
	if err := diagnostics.SummaryAsError(e.Diagnostics); err != nil {
		return err.Error()
	}
	return "syntax error"
}

// ErrUnstable means the formatter produced output whose tokens differ from
// the input's. It indicates a bug in the formatter; the output is discarded.
var ErrUnstable = errors.New("formatter output does not preserve the program's tokens")

// Format returns src laid out canonically.
func Format(src []rune, opts Options) (string, error) {
	if opts.Indent == "" {
		opts.Indent = "\t"
	}

	if err := checkSyntax(src); err != nil {
		return "", err
	}

	tokens := tokenizeWithComments(src)
	out := render(tokens, src, opts)

	if !sameTokens(tokens, src, tokenizeWithComments([]rune(out)), []rune(out)) {
		return "", ErrUnstable
	}

	if strings.Contains(string(src), "\r\n") {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	return out, nil
}

func checkSyntax(src []rune) error {
	lex := lexer.NewLexer(src)
	tokens, _ := lex.Tokenize()
	if ds := lex.Diagnostics(); len(ds) > 0 {
		return &SyntaxError{Diagnostics: ds}
	}

	p := parser.Parser{Tokens: tokens}
	if _, ds := p.Parse(); diagnostics.Summary(ds).AnyError() {
		return &SyntaxError{Diagnostics: ds}
	}
	return nil
}

func tokenizeWithComments(src []rune) []lexer.Token {
	lex := lexer.NewLexer(src)
	lex.KeepComments = true
	tokens, _ := lex.Tokenize()
	return tokens
}

// render lays out the tokens line by line. Every source line that has
// tokens becomes exactly one output line, so the line structure the parser
// relies on is unchanged.
func render(tokens []lexer.Token, src []rune, opts Options) string {
	var (
		lines   []string
		current []lexer.Token
		blocks  indenter
	)

	flush := func() {
		if len(current) == 0 {
			lines = append(lines, "")
			return
		}
		glyphs := buildGlyphs(current, src)
		level := blocks.line(glyphs)
		lines = append(lines, strings.Repeat(opts.Indent, level)+join(glyphs))
		current = current[:0]
	}

	for _, tok := range tokens {
		if tok.Type == lexer.NEWLINE {
			flush()
			continue
		}
		current = append(current, tok)
	}
	if len(current) > 0 {
		flush()
	}

	var b strings.Builder
	blank := false
	for _, line := range lines {
		if line == "" {
			blank = b.Len() > 0
			continue
		}
		if blank {
			b.WriteByte('\n')
			blank = false
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

// sameTokens reports whether two token streams are equivalent for the
// formatter: same tokens with the same values (keywords and logical
// literals compared without case), and newlines in the same places, with
// runs of blank lines treated as one.
func sameTokens(a []lexer.Token, srcA []rune, b []lexer.Token, srcB []rune) bool {
	keysA, keysB := tokenKeys(a, srcA), tokenKeys(b, srcB)
	if len(keysA) != len(keysB) {
		return false
	}
	for i := range keysA {
		if keysA[i] != keysB[i] {
			return false
		}
	}
	return true
}

func tokenKeys(tokens []lexer.Token, src []rune) []string {
	const newline = "\n"
	var keys []string
	afterComo := false
	for _, tok := range tokens {
		switch {
		case tok.Type == lexer.NEWLINE:
			if len(keys) > 0 && keys[len(keys)-1] != newline {
				keys = append(keys, newline)
			}
		case tok.Type == lexer.KEYWORD, tok.Type == lexer.BOOLEAN_LITERAL,
			// Type names are matched without case by the analyzer.
			tok.Type == lexer.IDENTIFIER && afterComo:
			keys = append(keys, tok.Type.String()+":"+strings.ToLower(tok.Value))
		default:
			keys = append(keys, tok.Type.String()+":"+tokenText(tok, src))
		}
		if tok.Type != lexer.COMMENT {
			kw, ok := lexer.MapToKeyword([]rune(tok.Value))
			afterComo = tok.Type == lexer.KEYWORD && ok && kw == lexer.COMO
		}
	}
	for len(keys) > 0 && keys[len(keys)-1] == newline {
		keys = keys[:len(keys)-1]
	}
	return keys
}
