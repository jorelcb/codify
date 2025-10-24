package template

import "fmt"

// TokenType represents the type of a token in the template
type TokenType int

const (
	// TokenText represents plain text content
	TokenText TokenType = iota
	// TokenVariable represents a variable like {{VAR}}
	TokenVariable
	// TokenOpenBlock represents an opening block like {{#if}}
	TokenOpenBlock
	// TokenCloseBlock represents a closing block like {{/if}}
	TokenCloseBlock
	// TokenIdentifier represents an identifier (variable name, function name, keyword)
	TokenIdentifier
	// TokenString represents a string literal like "api"
	TokenString
	// TokenLParen represents a left parenthesis
	TokenLParen
	// TokenRParen represents a right parenthesis
	TokenRParen
	// TokenEOF represents end of file
	TokenEOF
	// TokenError represents an error
	TokenError
)

// String returns string representation of token type
func (t TokenType) String() string {
	switch t {
	case TokenText:
		return "TEXT"
	case TokenVariable:
		return "VARIABLE"
	case TokenOpenBlock:
		return "OPEN_BLOCK"
	case TokenCloseBlock:
		return "CLOSE_BLOCK"
	case TokenIdentifier:
		return "IDENTIFIER"
	case TokenString:
		return "STRING"
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Token represents a single token in the template
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

// String returns string representation of token
func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %d:%d", t.Type, t.Value, t.Line, t.Col)
}