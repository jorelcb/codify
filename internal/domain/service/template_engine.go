package service

import (
	"fmt"
	"strings"

	"github.com/jorelcb/ai-context-generator/internal/domain/template"
)

// TemplateEngine renders templates by evaluating AST with context
type TemplateEngine struct {
	helpers map[string]HelperFunc
}

// HelperFunc is a function that can be called from templates
type HelperFunc func(args []interface{}) (bool, error)

// NewTemplateEngine creates a new template engine with default helpers
func NewTemplateEngine() *TemplateEngine {
	engine := &TemplateEngine{
		helpers: make(map[string]HelperFunc),
	}

	// Register default helpers
	engine.RegisterHelper("eq", helperEq)
	engine.RegisterHelper("includes", helperIncludes)

	return engine
}

// RegisterHelper registers a helper function
func (e *TemplateEngine) RegisterHelper(name string, fn HelperFunc) {
	e.helpers[name] = fn
}

// Render renders a template with the given context
func (e *TemplateEngine) Render(input string, context map[string]interface{}) (string, error) {
	// Parse the template
	ast, err := template.Parse(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Render the AST
	return e.renderNode(ast, context)
}

// renderNode renders a single AST node
func (e *TemplateEngine) renderNode(node template.Node, context map[string]interface{}) (string, error) {
	switch n := node.(type) {
	case *template.RootNode:
		return e.renderRoot(n, context)
	case *template.TextNode:
		return n.Content, nil
	case *template.VariableNode:
		return e.renderVariable(n, context)
	case *template.IfBlockNode:
		return e.renderIfBlock(n, context)
	case *template.EachBlockNode:
		return e.renderEachBlock(n, context)
	default:
		return "", fmt.Errorf("unknown node type: %T", node)
	}
}

// renderRoot renders a root node
func (e *TemplateEngine) renderRoot(node *template.RootNode, context map[string]interface{}) (string, error) {
	var result strings.Builder

	for _, child := range node.Children {
		rendered, err := e.renderNode(child, context)
		if err != nil {
			return "", err
		}
		result.WriteString(rendered)
	}

	return result.String(), nil
}

// renderVariable renders a variable node
func (e *TemplateEngine) renderVariable(node *template.VariableNode, context map[string]interface{}) (string, error) {
	value, exists := context[node.Name]
	if !exists {
		return "", fmt.Errorf("variable %s not found in context", node.Name)
	}

	return fmt.Sprintf("%v", value), nil
}

// renderIfBlock renders an if block
func (e *TemplateEngine) renderIfBlock(node *template.IfBlockNode, context map[string]interface{}) (string, error) {
	// Evaluate condition
	condition, err := e.evaluateCondition(node.Condition, context)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate if condition: %w", err)
	}

	// Render then or else body
	var nodes []template.Node
	if condition {
		nodes = node.ThenBody
	} else {
		nodes = node.ElseBody
	}

	var result strings.Builder
	for _, child := range nodes {
		rendered, err := e.renderNode(child, context)
		if err != nil {
			return "", err
		}
		result.WriteString(rendered)
	}

	return result.String(), nil
}

// renderEachBlock renders an each block
func (e *TemplateEngine) renderEachBlock(node *template.EachBlockNode, context map[string]interface{}) (string, error) {
	// Get the array from context
	arrayValue, exists := context[node.ArrayName]
	if !exists {
		return "", fmt.Errorf("array %s not found in context", node.ArrayName)
	}

	// Convert to slice
	array, ok := arrayValue.([]interface{})
	if !ok {
		// Try to convert from []string
		if strArray, ok := arrayValue.([]string); ok {
			array = make([]interface{}, len(strArray))
			for i, s := range strArray {
				array[i] = s
			}
		} else {
			return "", fmt.Errorf("variable %s is not an array", node.ArrayName)
		}
	}

	// Render body for each item
	var result strings.Builder
	for _, item := range array {
		// Create new context with current item
		itemContext := make(map[string]interface{})
		for k, v := range context {
			itemContext[k] = v
		}
		itemContext[node.ItemName] = item
		itemContext["this"] = item

		// Render body with item context
		for _, child := range node.Body {
			rendered, err := e.renderNode(child, itemContext)
			if err != nil {
				return "", err
			}
			result.WriteString(rendered)
		}
	}

	return result.String(), nil
}

// evaluateCondition evaluates a condition node to a boolean
func (e *TemplateEngine) evaluateCondition(node template.Node, context map[string]interface{}) (bool, error) {
	switch n := node.(type) {
	case *template.VariableNode:
		// Simple variable truthiness
		value, exists := context[n.Name]
		if !exists {
			return false, nil
		}
		return isTruthy(value), nil

	case *template.ExpressionNode:
		// Evaluate helper function
		return e.evaluateExpression(n, context)

	default:
		return false, fmt.Errorf("cannot evaluate condition of type %T", node)
	}
}

// evaluateExpression evaluates an expression node
func (e *TemplateEngine) evaluateExpression(node *template.ExpressionNode, context map[string]interface{}) (bool, error) {
	helper, exists := e.helpers[node.Function]
	if !exists {
		return false, fmt.Errorf("unknown helper function: %s", node.Function)
	}

	// Resolve arguments
	args := make([]interface{}, len(node.Args))
	for i, arg := range node.Args {
		args[i] = e.resolveArgument(arg, context)
	}

	// Call helper
	return helper(args)
}

// resolveArgument resolves an argument (variable or literal)
func (e *TemplateEngine) resolveArgument(arg string, context map[string]interface{}) interface{} {
	// Check if it's a variable in context
	if value, exists := context[arg]; exists {
		return value
	}

	// Otherwise treat as literal string
	return arg
}

// isTruthy checks if a value is truthy
func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int, int32, int64:
		return v != 0
	case []interface{}:
		return len(v) > 0
	case []string:
		return len(v) > 0
	default:
		return true
	}
}

// Helper functions

// helperEq checks if two values are equal
func helperEq(args []interface{}) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("eq requires exactly 2 arguments, got %d", len(args))
	}

	return fmt.Sprintf("%v", args[0]) == fmt.Sprintf("%v", args[1]), nil
}

// helperIncludes checks if an array includes a value
func helperIncludes(args []interface{}) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("includes requires exactly 2 arguments, got %d", len(args))
	}

	// Convert first arg to array
	var array []interface{}
	switch v := args[0].(type) {
	case []interface{}:
		array = v
	case []string:
		array = make([]interface{}, len(v))
		for i, s := range v {
			array[i] = s
		}
	default:
		return false, fmt.Errorf("first argument to includes must be an array")
	}

	// Check if array includes the value
	searchValue := fmt.Sprintf("%v", args[1])
	for _, item := range array {
		if fmt.Sprintf("%v", item) == searchValue {
			return true, nil
		}
	}

	return false, nil
}