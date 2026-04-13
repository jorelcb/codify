package catalog

import (
	"regexp"
	"strconv"
	"strings"
)

// AnnotationMeta represents a parsed Antigravity execution annotation from a workflow template.
type AnnotationMeta struct {
	Type     string // "turbo", "capture", "if"
	Value    string // "" for turbo, variable name for capture, condition text for if
	Step     int    // Step number where annotation appears
	StepName string // Step title (e.g., "Run Full Test Suite")
}

var stepHeaderRegex = regexp.MustCompile(`^###\s+(\d+)\.\s+(.+)$`)

// ParseAnnotations scans a workflow template and extracts structured annotation metadata.
func ParseAnnotations(templateContent string) []AnnotationMeta {
	var annotations []AnnotationMeta
	var currentStep int
	var currentStepName string

	lines := strings.Split(templateContent, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track current step from ### N. headers
		if matches := stepHeaderRegex.FindStringSubmatch(trimmed); matches != nil {
			n, err := strconv.Atoi(matches[1])
			if err == nil {
				currentStep = n
				currentStepName = strings.TrimSpace(matches[2])
			}
			continue
		}

		// Parse annotations
		if strings.HasPrefix(trimmed, "// turbo") {
			annotations = append(annotations, AnnotationMeta{
				Type:     "turbo",
				Step:     currentStep,
				StepName: currentStepName,
			})
		} else if strings.HasPrefix(trimmed, "// capture:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "// capture:"))
			annotations = append(annotations, AnnotationMeta{
				Type:     "capture",
				Value:    value,
				Step:     currentStep,
				StepName: currentStepName,
			})
		} else if strings.HasPrefix(trimmed, "// if ") {
			condition := strings.TrimSpace(strings.TrimPrefix(trimmed, "// if "))
			annotations = append(annotations, AnnotationMeta{
				Type:     "if",
				Value:    condition,
				Step:     currentStep,
				StepName: currentStepName,
			})
		} else if strings.HasPrefix(trimmed, "// parallel") {
			annotations = append(annotations, AnnotationMeta{
				Type:     "parallel",
				Step:     currentStep,
				StepName: currentStepName,
			})
		} else if strings.HasPrefix(trimmed, "// retry:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "// retry:"))
			annotations = append(annotations, AnnotationMeta{
				Type:     "retry",
				Value:    value,
				Step:     currentStep,
				StepName: currentStepName,
			})
		} else if strings.HasPrefix(trimmed, "// timeout:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "// timeout:"))
			annotations = append(annotations, AnnotationMeta{
				Type:     "timeout",
				Value:    value,
				Step:     currentStep,
				StepName: currentStepName,
			})
		}
	}

	return annotations
}

// FilterByType returns annotations of a specific type.
func FilterByType(annotations []AnnotationMeta, annotationType string) []AnnotationMeta {
	var filtered []AnnotationMeta
	for _, a := range annotations {
		if a.Type == annotationType {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

// HasAnnotationType returns true if any annotation of the given type exists.
func HasAnnotationType(annotations []AnnotationMeta, annotationType string) bool {
	for _, a := range annotations {
		if a.Type == annotationType {
			return true
		}
	}
	return false
}