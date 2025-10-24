package service

import (
	"strings"
	"testing"
)

func TestTemplateEngine_SimpleVariable(t *testing.T) {
	engine := NewTemplateEngine()
	context := map[string]interface{}{
		"PROJECT_NAME": "MyApp",
	}

	result, err := engine.Render("Hello {{PROJECT_NAME}}!", context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Hello MyApp!"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTemplateEngine_IfBlockWithEq(t *testing.T) {
	engine := NewTemplateEngine()
	context := map[string]interface{}{
		"PROJECT_TYPE": "api",
	}

	template := `{{#if (eq PROJECT_TYPE "api")}}This is an API{{/if}}`
	result, err := engine.Render(template, context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "This is an API"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTemplateEngine_IfBlockFalse(t *testing.T) {
	engine := NewTemplateEngine()
	context := map[string]interface{}{
		"PROJECT_TYPE": "cli",
	}

	template := `{{#if (eq PROJECT_TYPE "api")}}This is an API{{/if}}`
	result, err := engine.Render(template, context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := ""
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTemplateEngine_EachBlock(t *testing.T) {
	engine := NewTemplateEngine()
	context := map[string]interface{}{
		"ITEMS": []string{"apple", "banana", "cherry"},
	}

	template := `{{#each ITEMS}}- {{this}}
{{/each}}`
	result, err := engine.Render(template, context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "- apple\n- banana\n- cherry\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTemplateEngine_IncludesHelper(t *testing.T) {
	engine := NewTemplateEngine()
	context := map[string]interface{}{
		"CAPABILITIES": []string{"messaging", "logging", "caching"},
	}

	template := `{{#if (includes CAPABILITIES "messaging")}}Has messaging{{/if}}`
	result, err := engine.Render(template, context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Has messaging"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTemplateEngine_ComplexTemplate(t *testing.T) {
	engine := NewTemplateEngine()
	context := map[string]interface{}{
		"PROJECT_NAME": "MyAPI",
		"PROJECT_TYPE": "api",
		"CAPABILITIES":  []string{"messaging", "logging"},
	}

	template := `# {{PROJECT_NAME}}

{{#if (eq PROJECT_TYPE "api")}}
## API Features
{{#each CAPABILITIES}}
- {{this}}
{{/each}}
{{/if}}`

	result, err := engine.Render(template, context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that result contains expected parts
	if !strings.Contains(result, "# MyAPI") {
		t.Error("result should contain project name")
	}
	if !strings.Contains(result, "## API Features") {
		t.Error("result should contain API Features section")
	}
	if !strings.Contains(result, "- messaging") {
		t.Error("result should contain messaging capability")
	}
	if !strings.Contains(result, "- logging") {
		t.Error("result should contain logging capability")
	}
}

func TestTemplateEngine_MissingVariable(t *testing.T) {
	engine := NewTemplateEngine()
	context := map[string]interface{}{}

	result, err := engine.Render("Hello {{PROJECT_NAME}}!", context)
	if err == nil {
		t.Fatal("expected error for missing variable")
	}

	if !strings.Contains(err.Error(), "not found in context") {
		t.Errorf("expected 'not found in context' error, got: %v", err)
	}

	if result != "" {
		t.Errorf("expected empty result on error, got %q", result)
	}
}

func TestTemplateEngine_UnknownHelper(t *testing.T) {
	engine := NewTemplateEngine()
	context := map[string]interface{}{
		"VAR": "value",
	}

	template := `{{#if (unknown VAR "value")}}test{{/if}}`
	_, err := engine.Render(template, context)
	if err == nil {
		t.Fatal("expected error for unknown helper")
	}

	if !strings.Contains(err.Error(), "unknown helper function") {
		t.Errorf("expected 'unknown helper function' error, got: %v", err)
	}
}