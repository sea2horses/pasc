package lexer

import (
	"pascc/internal/diagnostics"
	"pascc/internal/models"
	"unicode"
)

/* Lexer struct */
type Lexer struct {
	Src      []rune /* Loaded source code */
	position models.Position
	diagnostics.DiagnosticEmitter
}

func NewLexer(src []rune) *Lexer {
	return &Lexer{
		Src:               src,
		DiagnosticEmitter: diagnostics.NewDiagnosticEmitter(),
		position: models.Position{
			Line:   1,
			Column: 1,
		},
	}
}

/* Get character at current position */
func (l *Lexer) getch() (rune, error) {
	return l.at(l.position.Offset)
}

/* Get character at any position */
func (l *Lexer) at(position uint32) (rune, error) {
	if int(position) >= len(l.Src) {
		return 0, ErrOutOfBounds{Max: uint32(len(l.Src)), Value: position}
	}

	return l.Src[position], nil
}

/* Advance function */
func (l *Lexer) advance() error {
	ch, err := l.getch()

	if err != nil {
		return err
	}

	/* If there is a line break, reset the column and add to the line in file */
	if ch == '\n' {
		l.position.Line++
		l.position.Column = 1
	} else {
		l.position.Column++
	}

	l.position.Offset++
	return nil
}

/* Emit a token */
func (l *Lexer) emmit(token_type TokenType, value string, span models.Span) Token {
	return Token{
		Value: value,
		Type:  token_type,
		Span:  span,
	}
}

/* Skip Whitespace */
func (l *Lexer) skip_whitespace() {
	for {
		ch, err := l.getch()

		if err != nil || ch == '\n' || !unicode.IsSpace(ch) {
			break
		}

		l.advance()
	}
}

/* Parse identifiers */
func (l *Lexer) tokenize_identifier() (Token, error) {
	init_position := l.position

	/* Run until reaching the end of the file or putting a stop */
	for {
		ch, err := l.getch()

		if err != nil || (!unicode.IsLetter(ch) && !unicode.IsNumber(ch) && ch != '_') {
			break
		}

		l.advance()
	}

	end_position := l.position

	init_ch, err := l.at(init_position.Offset)

	if err != nil {
		return NULL_TOKEN, err
	}

	/* If first char is NOT an identifier friendly character, throw an error */
	if !unicode.IsLetter(init_ch) && init_ch != '_' {
		l.Report(
			models.Span{Start: init_position, End: init_position},
			"%s",
			ErrExpectedIdentifier.Error(),
		)

		return NULL_TOKEN, ErrExpectedIdentifier
	}

	/* Else, emmit the token */
	value := l.Src[init_position.Offset:end_position.Offset]

	_, ok := MapToKeyword(value)
	token_type := IDENTIFIER

	if ok {
		token_type = KEYWORD
	}

	/* Try parse bool literal */
	if !ok {
		_, ok := MapToBool(value)
		if ok {
			token_type = BOOLEAN_LITERAL
		}
	}

	return l.emmit(token_type, string(value), models.Span{Start: init_position, End: end_position}), nil
}

/* Parse string literal */
func (l *Lexer) tokenize_string_literal() (Token, error) {
	init_position := l.position

	/* Skip first " */
	l.advance()

	/* Run until reaching the end of the file or putting a stop */
	for {
		ch, err := l.getch()

		if err != nil || ch == '"' {
			break
		}

		l.advance()
	}

	end_position := l.position

	/* Check that the init position and the end poisiton exist */
	init_ch, err := l.at(init_position.Offset)

	if err != nil {
		return NULL_TOKEN, err
	}

	end_ch, err := l.at(end_position.Offset)

	if err != nil {
		l.Report(
			models.Span{Start: init_position, End: init_position},
			"%s",
			ErrUnterminatedString.Error(),
		)

		return NULL_TOKEN, ErrUnterminatedString
	}

	/* Skip last '"' */
	l.advance()

	/* Check that both ends of our string are good */
	if init_ch != '"' {
		l.Report(
			models.Span{Start: init_position, End: init_position},
			"%s",
			ErrExpectedString.Error(),
		)

		return NULL_TOKEN, ErrExpectedString
	}

	if end_ch != '"' {
		l.Report(
			models.Span{Start: init_position, End: init_position},
			"%s",
			ErrUnterminatedString.Error(),
		)

		return NULL_TOKEN, ErrUnterminatedString
	}

	/* Emit string token */
	value := l.Src[init_position.Offset+1 : end_position.Offset]

	return l.emmit(STRING_LITERAL, string(value), models.Span{Start: init_position, End: end_position}), nil
}

/* Parse number literal */
func (l *Lexer) tokenize_number_literal() (Token, error) {
	init_position := l.position

	for {
		ch, err := l.getch()

		if err != nil || (!unicode.IsNumber(ch)) {
			break
		}

		l.advance()
	}

	end_position := l.position

	init_ch, err := l.at(init_position.Offset)
	if err != nil {
		return NULL_TOKEN, err
	}

	/* Number literals can't start with . */
	if !unicode.IsNumber(init_ch) {
		return NULL_TOKEN, ErrExpectedNumber
	}

	value := l.Src[init_position.Offset:end_position.Offset]
	_, ok := MapToNumber(value)

	if !ok {
		l.Report(
			models.Span{Start: init_position, End: end_position},
			"%s",
			ErrExpectedNumber.Error(),
		)

		return NULL_TOKEN, ErrExpectedNumber
	}

	return l.emmit(NUMBER_LITERAL, string(value), models.Span{Start: init_position, End: end_position}), nil
}

/* TODO: Parse char */
func (l *Lexer) tokenize_character() (Token, error) {
	/* Get current char */
	ch, err := l.getch()

	if err != nil {
		return NULL_TOKEN, err
	}

	capture_position := l.position

	l.advance()

	tk_type, ok := charMap[ch]
	if !ok {
		l.Report(
			models.Span{Start: capture_position, End: capture_position},
			"%s",
			ErrUnexpectedChar{Char: ch}.Error(),
		)

		return NULL_TOKEN, ErrUnexpectedChar{Char: ch}
	}

	return Token{Type: tk_type, Value: string(ch), Span: models.Span{Start: capture_position, End: capture_position}}, nil
}

/* TODO: Use newline to better and more intelligently report diagnostics */
/* Tokenize function */
func (l *Lexer) Tokenize() ([]Token, error) {
	/* Make empty token list */
	tokens := []Token{}

	/* While there is source code to parse */
	for l.position.Offset < uint32(len(l.Src)) {
		ch, _ := l.getch()

		if unicode.IsSpace(ch) && ch != '\n' {
			l.skip_whitespace()
		} else if ch == '"' {
			token, err := l.tokenize_string_literal()
			if err == nil {
				tokens = append(tokens, token)
			}
		} else if unicode.IsLetter(ch) || ch == '_' {
			token, err := l.tokenize_identifier()
			if err == nil {
				tokens = append(tokens, token)
			}
		} else if unicode.IsNumber(ch) {
			token, err := l.tokenize_number_literal()
			if err == nil {
				tokens = append(tokens, token)
			}
		} else {
			token, err := l.tokenize_character()
			if err == nil {
				tokens = append(tokens, token)
			}
		}
	}

	return tokens, nil
}
