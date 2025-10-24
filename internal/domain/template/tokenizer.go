package template

import (
	"fmt"
	"strings"
	"unicode"
)

// Tokenizer scans template content and produces tokens
type Tokenizer struct {
	input   string
	pos     int  // current position
	readPos int  // reading position (after current char)
	ch      byte // current char
	line    int
	col     int
}

// NewTokenizer creates a new tokenizer for the given input
func NewTokenizer(input string) *Tokenizer {
	t := &Tokenizer{
		input: input,
		line:  1,
		col:   0,
	}
	t.readChar()
	return t
}

// readChar reads the next character
func (t *Tokenizer) readChar() {
	if t.readPos >= len(t.input) {
		t.ch = 0 // EOF
	} else {
		t.ch = t.input[t.readPos]
	}
	t.pos = t.readPos
	t.readPos++
	t.col++

	if t.ch == '\n' {
		t.line++
		t.col = 0
	}
}

// peekChar returns the next character without advancing
func (t *Tokenizer) peekChar() byte {
	if t.readPos >= len(t.input) {
		return 0
	}
	return t.input[t.readPos]
}

// NextToken returns the next token from the input
func (t *Tokenizer) NextToken() Token {
	// Check for template delimiters {{ }}
	if t.ch == '{' && t.peekChar() == '{' {
		return t.readTemplateTag()
	}

	// Read plain text until we hit {{ or EOF
	return t.readText()
}

// readText reads plain text until {{ or EOF
func (t *Tokenizer) readText() Token {
	startLine := t.line
	startCol := t.col
	var text strings.Builder

	for t.ch != 0 {
		// Check if we're at the start of a template tag
		if t.ch == '{' && t.peekChar() == '{' {
			break
		}
		text.WriteByte(t.ch)
		t.readChar()
	}

	value := text.String()
	if value == "" {
		return Token{Type: TokenEOF, Line: t.line, Col: t.col}
	}

	return Token{
		Type:  TokenText,
		Value: value,
		Line:  startLine,
		Col:   startCol,
	}
}

// readTemplateTag reads a template tag starting with {{
func (t *Tokenizer) readTemplateTag() Token {
	startLine := t.line
	startCol := t.col

	// Consume {{
	t.readChar()
	t.readChar()

	// Skip whitespace
	t.skipWhitespace()

	// Check for block markers # or /
	if t.ch == '#' {
		t.readChar()
		t.skipWhitespace()
		keyword := t.readIdentifier()
		content := t.readTagContent()
		t.skipToCloseBrace() // Consume the }}
		return Token{
			Type:  TokenOpenBlock,
			Value: keyword + content,
			Line:  startLine,
			Col:   startCol,
		}
	}

	if t.ch == '/' {
		t.readChar()
		t.skipWhitespace()
		keyword := t.readIdentifier()
		t.skipToCloseBrace()
		return Token{
			Type:  TokenCloseBlock,
			Value: keyword,
			Line:  startLine,
			Col:   startCol,
		}
	}

	// Simple variable
	variable := t.readIdentifier()
	t.skipToCloseBrace()
	return Token{
		Type:  TokenVariable,
		Value: variable,
		Line:  startLine,
		Col:   startCol,
	}
}

// readTagContent reads everything until }}
func (t *Tokenizer) readTagContent() string {
	var content strings.Builder
	t.skipWhitespace()

	for t.ch != 0 {
		if t.ch == '}' && t.peekChar() == '}' {
			break
		}
		content.WriteByte(t.ch)
		t.readChar()
	}

	return strings.TrimSpace(content.String())
}

// readIdentifier reads an identifier (letters, numbers, underscores)
func (t *Tokenizer) readIdentifier() string {
	var ident strings.Builder
	for isIdentifierChar(t.ch) {
		ident.WriteByte(t.ch)
		t.readChar()
	}
	return ident.String()
}

// skipToCloseBrace skips everything until }}
func (t *Tokenizer) skipToCloseBrace() {
	for t.ch != 0 {
		if t.ch == '}' && t.peekChar() == '}' {
			t.readChar() // consume first }
			t.readChar() // consume second }
			return
		}
		t.readChar()
	}
}

// skipWhitespace skips whitespace characters
func (t *Tokenizer) skipWhitespace() {
	for t.ch == ' ' || t.ch == '\t' || t.ch == '\n' || t.ch == '\r' {
		t.readChar()
	}
}

// isIdentifierChar returns true if the character is valid in an identifier
func isIdentifierChar(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_'
}

// Tokenize tokenizes the entire input and returns all tokens
func Tokenize(input string) ([]Token, error) {
	tokenizer := NewTokenizer(input)
	var tokens []Token

	for {
		token := tokenizer.NextToken()
		tokens = append(tokens, token)

		if token.Type == TokenEOF {
			break
		}

		if token.Type == TokenError {
			return nil, fmt.Errorf("tokenization error: %s", token.Value)
		}
	}

	return tokens, nil
}