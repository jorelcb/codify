package catalog

import (
	"testing"
)

func TestParseAnnotations_Turbo(t *testing.T) {
	content := `### 3. Update Version References
// turbo
Update version strings in project files.`

	annotations := ParseAnnotations(content)
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}
	if annotations[0].Type != "turbo" {
		t.Errorf("expected type 'turbo', got '%s'", annotations[0].Type)
	}
	if annotations[0].Step != 3 {
		t.Errorf("expected step 3, got %d", annotations[0].Step)
	}
	if annotations[0].StepName != "Update Version References" {
		t.Errorf("expected step name 'Update Version References', got '%s'", annotations[0].StepName)
	}
}

func TestParseAnnotations_Capture(t *testing.T) {
	content := `### 2. Determine Version Number
// capture: NEW_VERSION
Determine the new version following Semantic Versioning.`

	annotations := ParseAnnotations(content)
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}
	if annotations[0].Type != "capture" {
		t.Errorf("expected type 'capture', got '%s'", annotations[0].Type)
	}
	if annotations[0].Value != "NEW_VERSION" {
		t.Errorf("expected value 'NEW_VERSION', got '%s'", annotations[0].Value)
	}
	if annotations[0].Step != 2 {
		t.Errorf("expected step 2, got %d", annotations[0].Step)
	}
}

func TestParseAnnotations_If(t *testing.T) {
	content := `### 8. Trigger Deployment
// if the project has CI/CD deployment
Initiate the deployment pipeline.`

	annotations := ParseAnnotations(content)
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}
	if annotations[0].Type != "if" {
		t.Errorf("expected type 'if', got '%s'", annotations[0].Type)
	}
	if annotations[0].Value != "the project has CI/CD deployment" {
		t.Errorf("expected condition text, got '%s'", annotations[0].Value)
	}
}

func TestParseAnnotations_Empty(t *testing.T) {
	content := `# A Workflow
## Purpose
Just a plain workflow with no annotations.

### 1. Do Something
Do something without annotations.`

	annotations := ParseAnnotations(content)
	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(annotations))
	}
}

func TestParseAnnotations_ReleaseCycleTemplate(t *testing.T) {
	// Mirrors the actual release_cycle.template content
	content := `# Release Cycle Workflow

### 1. Verify Release Readiness
Before starting the release.

### 2. Determine Version Number
// capture: NEW_VERSION
Determine the new version.

### 3. Update Version References
// turbo
Update version strings.

### 4. Generate Changelog
Create or update the changelog.

### 5. Create Release Commit
// turbo
Create a single commit.

### 6. Create Git Tag
// turbo
Tag the release commit.

### 7. Push Release
// turbo
Push the release to the remote.

### 8. Trigger Deployment
// if the project has CI/CD deployment
Initiate the deployment pipeline.

### 9. Create Release Notes
// if using GitHub/GitLab releases
Create a release on the hosting platform.

### 10. Post-Release Verification
// turbo
Verify the release is healthy.`

	annotations := ParseAnnotations(content)

	turbo := FilterByType(annotations, "turbo")
	capture := FilterByType(annotations, "capture")
	ifs := FilterByType(annotations, "if")

	if len(turbo) != 5 {
		t.Errorf("expected 5 turbo annotations, got %d", len(turbo))
	}
	if len(capture) != 1 {
		t.Errorf("expected 1 capture annotation, got %d", len(capture))
	}
	if len(ifs) != 2 {
		t.Errorf("expected 2 if annotations, got %d", len(ifs))
	}
	if capture[0].Value != "NEW_VERSION" {
		t.Errorf("expected capture value 'NEW_VERSION', got '%s'", capture[0].Value)
	}
}

func TestParseAnnotations_FeatureDevelopment(t *testing.T) {
	content := `### 1. Create Feature Branch
// capture: BRANCH_NAME
Create a new branch.

### 3. Plan the Implementation
// if the change touches more than 3 files
Break down the feature.

### 5. Run Full Test Suite
// turbo
Run the complete test suite.

### 8. Address Review Feedback
// if there is review feedback
Process each review comment.

### 9. Merge and Clean Up
// turbo
After approval.`

	annotations := ParseAnnotations(content)

	if len(annotations) != 5 {
		t.Fatalf("expected 5 annotations, got %d", len(annotations))
	}

	turbo := FilterByType(annotations, "turbo")
	capture := FilterByType(annotations, "capture")
	ifs := FilterByType(annotations, "if")

	if len(turbo) != 2 {
		t.Errorf("expected 2 turbo, got %d", len(turbo))
	}
	if len(capture) != 1 {
		t.Errorf("expected 1 capture, got %d", len(capture))
	}
	if len(ifs) != 2 {
		t.Errorf("expected 2 if, got %d", len(ifs))
	}
	if capture[0].Value != "BRANCH_NAME" {
		t.Errorf("expected 'BRANCH_NAME', got '%s'", capture[0].Value)
	}
}

func TestFilterByType(t *testing.T) {
	annotations := []AnnotationMeta{
		{Type: "turbo", Step: 1},
		{Type: "capture", Step: 2, Value: "VAR"},
		{Type: "turbo", Step: 3},
		{Type: "if", Step: 4, Value: "condition"},
	}

	turbo := FilterByType(annotations, "turbo")
	if len(turbo) != 2 {
		t.Errorf("expected 2 turbo, got %d", len(turbo))
	}

	capture := FilterByType(annotations, "capture")
	if len(capture) != 1 {
		t.Errorf("expected 1 capture, got %d", len(capture))
	}
}

func TestHasAnnotationType(t *testing.T) {
	annotations := []AnnotationMeta{
		{Type: "turbo"},
		{Type: "capture", Value: "VAR"},
	}

	if !HasAnnotationType(annotations, "turbo") {
		t.Error("expected HasAnnotationType('turbo') to be true")
	}
	if !HasAnnotationType(annotations, "capture") {
		t.Error("expected HasAnnotationType('capture') to be true")
	}
	if HasAnnotationType(annotations, "if") {
		t.Error("expected HasAnnotationType('if') to be false")
	}
}