package template

import (
	"fmt"
	"strings"
)

// NodeType represents the type of an AST node
type NodeType int

const (
	// NodeRoot represents the root of the template
	NodeRoot NodeType = iota
	// NodeText represents plain text
	NodeText
	// NodeVariable represents a variable reference like {{VAR}}
	NodeVariable
	// NodeIfBlock represents an if conditional block
	NodeIfBlock
	// NodeEachBlock represents an each loop block
	NodeEachBlock
	// NodeExpression represents an expression like (eq a b)
	NodeExpression
)

// String returns string representation of node type
func (n NodeType) String() string {
	switch n {
	case NodeRoot:
		return "ROOT"
	case NodeText:
		return "TEXT"
	case NodeVariable:
		return "VARIABLE"
	case NodeIfBlock:
		return "IF_BLOCK"
	case NodeEachBlock:
		return "EACH_BLOCK"
	case NodeExpression:
		return "EXPRESSION"
	default:
		return "UNKNOWN"
	}
}

// Node is the interface for all AST nodes
type Node interface {
	Type() NodeType
	String() string
}

// RootNode represents the root of the template AST
type RootNode struct {
	Children []Node
}

// Type returns the node type
func (n *RootNode) Type() NodeType {
	return NodeRoot
}

// String returns string representation
func (n *RootNode) String() string {
	var parts []string
	for _, child := range n.Children {
		parts = append(parts, child.String())
	}
	return fmt.Sprintf("Root[%s]", strings.Join(parts, ", "))
}

// TextNode represents plain text content
type TextNode struct {
	Content string
}

// Type returns the node type
func (n *TextNode) Type() NodeType {
	return NodeText
}

// String returns string representation
func (n *TextNode) String() string {
	// Truncate long text for readability
	content := n.Content
	if len(content) > 50 {
		content = content[:50] + "..."
	}
	return fmt.Sprintf("Text(%q)", content)
}

// VariableNode represents a variable reference
type VariableNode struct {
	Name string
}

// Type returns the node type
func (n *VariableNode) Type() NodeType {
	return NodeVariable
}

// String returns string representation
func (n *VariableNode) String() string {
	return fmt.Sprintf("Variable(%s)", n.Name)
}

// IfBlockNode represents an if conditional block
type IfBlockNode struct {
	Condition Node
	ThenBody  []Node
	ElseBody  []Node // Optional else block
}

// Type returns the node type
func (n *IfBlockNode) Type() NodeType {
	return NodeIfBlock
}

// String returns string representation
func (n *IfBlockNode) String() string {
	result := fmt.Sprintf("If(condition: %s, then: [%d nodes]", n.Condition, len(n.ThenBody))
	if len(n.ElseBody) > 0 {
		result += fmt.Sprintf(", else: [%d nodes]", len(n.ElseBody))
	}
	result += ")"
	return result
}

// EachBlockNode represents an each loop block
type EachBlockNode struct {
	ArrayName string
	ItemName  string // The name to use for current item (default: "this")
	Body      []Node
}

// Type returns the node type
func (n *EachBlockNode) Type() NodeType {
	return NodeEachBlock
}

// String returns string representation
func (n *EachBlockNode) String() string {
	return fmt.Sprintf("Each(array: %s, item: %s, body: [%d nodes])", n.ArrayName, n.ItemName, len(n.Body))
}

// ExpressionNode represents an expression like (eq a b) or (includes array item)
type ExpressionNode struct {
	Function string   // Function name: "eq", "includes", etc.
	Args     []string // Arguments to the function
}

// Type returns the node type
func (n *ExpressionNode) Type() NodeType {
	return NodeExpression
}

// String returns string representation
func (n *ExpressionNode) String() string {
	return fmt.Sprintf("Expr(%s %v)", n.Function, n.Args)
}

// ParseExpression parses an expression string like "(eq PROJECT_TYPE \"api\")"
// Returns an ExpressionNode or nil if not a valid expression
func ParseExpression(expr string) (*ExpressionNode, error) {
	expr = strings.TrimSpace(expr)

	// Check if it starts with ( and ends with )
	if !strings.HasPrefix(expr, "(") || !strings.HasSuffix(expr, ")") {
		return nil, fmt.Errorf("expression must be wrapped in parentheses: %s", expr)
	}

	// Remove outer parentheses
	expr = strings.TrimSpace(expr[1 : len(expr)-1])

	// Split into tokens (function name and arguments)
	parts := splitExpressionParts(expr)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty expression")
	}

	return &ExpressionNode{
		Function: parts[0],
		Args:     parts[1:],
	}, nil
}

// splitExpressionParts splits an expression into parts, respecting quoted strings
func splitExpressionParts(expr string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(expr); i++ {
		ch := expr[i]

		switch {
		case ch == '"' || ch == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuotes = false
				quoteChar = 0
				// Add the quoted string without quotes
				parts = append(parts, current.String())
				current.Reset()
				continue
			}
			if inQuotes && ch == quoteChar {
				continue
			}

		case ch == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue

		default:
			current.WriteByte(ch)
		}
	}

	// Add last part if any
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}