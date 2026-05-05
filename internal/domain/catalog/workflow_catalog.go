package catalog

import (
	"fmt"
	"maps"
	"regexp"
	"strings"
)

// WorkflowMeta contains metadata for a workflow for static frontmatter generation.
type WorkflowMeta struct {
	Description string // max 250 chars (Antigravity constraint)
}

// WorkflowMetadata maps guide names to their metadata for frontmatter.
var WorkflowMetadata = map[string]WorkflowMeta{
	"bug_fix":       {Description: "Structured bug fix: reproduce, diagnose, fix, test, and PR"},
	"release_cycle": {Description: "Release process: version bump, changelog, tag, and deploy"},
	"spec_propose":  {Description: "Propose a feature change with spec deltas, design, tasks, and feature branch"},
	"spec_apply":    {Description: "Apply a spec-driven change: implement tasks, test, and create PR"},
	"spec_archive":  {Description: "Archive a completed change: merge spec deltas, finalize, and clean up"},
}

// GenerateWorkflowFrontmatter generates YAML frontmatter for a workflow based on the target ecosystem.
func GenerateWorkflowFrontmatter(guideName, target string) string {
	name := strings.ReplaceAll(guideName, "_", "-")
	meta, ok := WorkflowMetadata[guideName]
	if !ok {
		meta = WorkflowMeta{Description: fmt.Sprintf("Workflow for %s", name)}
	}

	switch target {
	case "antigravity":
		return fmt.Sprintf("---\ndescription: %s\n---\n", meta.Description)
	default: // claude
		return fmt.Sprintf("---\nname: %s\ndescription: %s\ndisable-model-invocation: true\nallowed-tools: Bash(*)\n---\n", name, meta.Description)
	}
}

// WorkflowCategories is the global registry of workflow categories.
var WorkflowCategories = []SkillCategory{
	{
		Name:      "workflows",
		Label:     "Workflows",
		Exclusive: false,
		Options: []SkillOption{
			{
				Name:        "bug-fix",
				Label:       "Bug Fix (reproduce → diagnose → fix → test → PR)",
				TemplateDir: "workflows",
				TemplateMapping: map[string]string{
					"bug_fix.template": "bug_fix",
				},
			},
			{
				Name:        "release-cycle",
				Label:       "Release Cycle (version → changelog → tag → deploy)",
				TemplateDir: "workflows",
				TemplateMapping: map[string]string{
					"release_cycle.template": "release_cycle",
				},
			},
			{
				Name:        "spec-driven-change",
				Label:       "Spec-driven Change (propose → apply → archive)",
				TemplateDir: "workflows",
				TemplateMapping: map[string]string{
					"spec_propose.template": "spec_propose",
					"spec_apply.template":   "spec_apply",
					"spec_archive.template": "spec_archive",
				},
			},
		},
	},
}

// WorkflowCategoryNames returns the names of all registered workflow categories.
func WorkflowCategoryNames() []string {
	names := make([]string, len(WorkflowCategories))
	for i, c := range WorkflowCategories {
		names[i] = c.Name
	}
	return names
}

// WorkflowPresetNames returns the names of all workflow presets across every
// workflow category, plus the "all" alias. Used for MCP enum validation.
func WorkflowPresetNames() []string {
	seen := map[string]bool{"all": true}
	names := []string{"all"}
	for _, c := range WorkflowCategories {
		for _, o := range c.Options {
			if seen[o.Name] {
				continue
			}
			seen[o.Name] = true
			names = append(names, o.Name)
		}
	}
	return names
}

// FindWorkflowCategory looks up a workflow category by name.
func FindWorkflowCategory(name string) (*SkillCategory, error) {
	for i := range WorkflowCategories {
		if WorkflowCategories[i].Name == name {
			return &WorkflowCategories[i], nil
		}
	}
	return nil, fmt.Errorf("unknown workflow category: %s", name)
}

// annotationLineRegex matches Antigravity execution annotation lines.
var annotationLineRegex = regexp.MustCompile(`^\s*//\s*(turbo|capture:|if |parallel|retry:|timeout:)`)

// StripAnnotationLines removes Antigravity execution annotation lines from workflow content.
// Non-annotation lines are preserved as-is.
func StripAnnotationLines(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		if annotationLineRegex.MatchString(strings.TrimSpace(line)) {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// ResolveAllWorkflows combines all workflow options into a single selection.
func ResolveAllWorkflows() *ResolvedSelection {
	merged := make(map[string]string)
	var dir string
	for _, cat := range WorkflowCategories {
		for _, opt := range cat.Options {
			dir = opt.TemplateDir
			maps.Copy(merged, opt.TemplateMapping)
		}
	}
	return &ResolvedSelection{
		TemplateDir:     dir,
		TemplateMapping: merged,
	}
}
