package template

import (
	"fmt"
	"strings"
)

// Parser parses tokens into an AST
type Parser struct {
	tokens  []Token
	pos     int
	current Token
}

// NewParser creates a new parser for the given tokens
func NewParser(tokens []Token) *Parser {
	p := &Parser{
		tokens: tokens,
		pos:    0,
	}
	if len(tokens) > 0 {
		p.current = tokens[0]
	}
	return p
}

// advance moves to the next token
func (p *Parser) advance() {
	p.pos++
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
	} else {
		p.current = Token{Type: TokenEOF}
	}
}

// peek returns the next token without advancing
func (p *Parser) peek() Token {
	if p.pos+1 < len(p.tokens) {
		return p.tokens[p.pos+1]
	}
	return Token{Type: TokenEOF}
}

// Parse parses the tokens into an AST
func (p *Parser) Parse() (*RootNode, error) {
	root := &RootNode{
		Children: make([]Node, 0),
	}

	for p.current.Type != TokenEOF {
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			root.Children = append(root.Children, node)
		}
	}

	return root, nil
}

// parseNode parses a single node based on current token
func (p *Parser) parseNode() (Node, error) {
	switch p.current.Type {
	case TokenText:
		node := &TextNode{Content: p.current.Value}
		p.advance()
		return node, nil

	case TokenVariable:
		node := &VariableNode{Name: p.current.Value}
		p.advance()
		return node, nil

	case TokenOpenBlock:
		return p.parseBlock()

	case TokenCloseBlock:
		// Close block should be handled by parseBlock
		return nil, fmt.Errorf("unexpected close block: %s", p.current.Value)

	case TokenEOF:
		return nil, nil

	default:
		return nil, fmt.Errorf("unexpected token type: %s", p.current.Type)
	}
}

// parseBlock parses a block (if or each)
func (p *Parser) parseBlock() (Node, error) {
	blockContent := p.current.Value
	p.advance()

	// Determine block type by looking at the keyword
	if strings.HasPrefix(blockContent, "if") {
		return p.parseIfBlock(blockContent)
	}

	if strings.HasPrefix(blockContent, "each") {
		return p.parseEachBlock(blockContent)
	}

	return nil, fmt.Errorf("unknown block type: %s", blockContent)
}

// parseIfBlock parses an if block
func (p *Parser) parseIfBlock(blockContent string) (*IfBlockNode, error) {
	// Extract condition from blockContent (after "if ")
	conditionStr := strings.TrimSpace(strings.TrimPrefix(blockContent, "if"))

	// Parse condition as expression or variable
	var condition Node
	if strings.HasPrefix(conditionStr, "(") {
		expr, err := ParseExpression(conditionStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse if condition: %w", err)
		}
		condition = expr
	} else {
		// Simple variable condition
		condition = &VariableNode{Name: conditionStr}
	}

	// Parse then body
	thenBody, err := p.parseUntilClose("if")
	if err != nil {
		return nil, err
	}

	return &IfBlockNode{
		Condition: condition,
		ThenBody:  thenBody,
	}, nil
}

// parseEachBlock parses an each block
func (p *Parser) parseEachBlock(blockContent string) (*EachBlockNode, error) {
	// Extract array name from blockContent (after "each ")
	arrayName := strings.TrimSpace(strings.TrimPrefix(blockContent, "each"))

	// Parse body
	body, err := p.parseUntilClose("each")
	if err != nil {
		return nil, err
	}

	return &EachBlockNode{
		ArrayName: arrayName,
		ItemName:  "this", // Default item name
		Body:      body,
	}, nil
}

// parseUntilClose parses nodes until a matching close block is found
func (p *Parser) parseUntilClose(blockName string) ([]Node, error) {
	var nodes []Node
	nestingLevel := 1

	for p.current.Type != TokenEOF {
		// Check for close block
		if p.current.Type == TokenCloseBlock {
			if p.current.Value == blockName {
				nestingLevel--
				if nestingLevel == 0 {
					p.advance() // consume close block
					return nodes, nil
				}
			}
			p.advance()
			continue
		}

		// Check for nested open block of same type
		if p.current.Type == TokenOpenBlock {
			if strings.HasPrefix(p.current.Value, blockName+" ") {
				nestingLevel++
			}
		}

		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	return nil, fmt.Errorf("unclosed block: %s", blockName)
}

// Parse is a convenience function to tokenize and parse in one step
func Parse(input string) (*RootNode, error) {
	tokens, err := Tokenize(input)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}

	parser := NewParser(tokens)
	return parser.Parse()
}
