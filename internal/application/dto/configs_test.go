package dto

import "testing"

func TestSkillsConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *SkillsConfig
		wantErr bool
	}{
		{
			"happy path",
			&SkillsConfig{OutputPath: "/tmp/x", Target: "claude"},
			false,
		},
		{
			"missing output",
			&SkillsConfig{Target: "claude"},
			true,
		},
		{
			"invalid target",
			&SkillsConfig{OutputPath: "/tmp/x", Target: "vscode"},
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestSpecConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *SpecConfig
		wantErr bool
	}{
		{"happy", &SpecConfig{ProjectName: "x", FromContextPath: "/c", OutputPath: "/o"}, false},
		{"missing name", &SpecConfig{FromContextPath: "/c", OutputPath: "/o"}, true},
		{"missing context path", &SpecConfig{ProjectName: "x", OutputPath: "/o"}, true},
		{"missing output", &SpecConfig{ProjectName: "x", FromContextPath: "/c"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestWorkflowConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *WorkflowConfig
		wantErr bool
	}{
		{
			"happy static",
			&WorkflowConfig{Category: "workflows", Preset: "all", Mode: "static", Target: "antigravity", OutputPath: "/o"},
			false,
		},
		{
			"happy personalized with context",
			&WorkflowConfig{Category: "workflows", Preset: "all", Mode: "personalized", Target: "claude", OutputPath: "/o", ProjectContext: "ctx"},
			false,
		},
		{
			"missing category",
			&WorkflowConfig{Preset: "all", Mode: "static", OutputPath: "/o"},
			true,
		},
		{
			"missing preset",
			&WorkflowConfig{Category: "workflows", Mode: "static", OutputPath: "/o"},
			true,
		},
		{
			"missing mode",
			&WorkflowConfig{Category: "workflows", Preset: "all", OutputPath: "/o"},
			true,
		},
		{
			"invalid mode",
			&WorkflowConfig{Category: "workflows", Preset: "all", Mode: "auto", OutputPath: "/o"},
			true,
		},
		{
			"invalid target",
			&WorkflowConfig{Category: "workflows", Preset: "all", Mode: "static", Target: "codex", OutputPath: "/o"},
			true,
		},
		{
			"personalized without context",
			&WorkflowConfig{Category: "workflows", Preset: "all", Mode: "personalized", Target: "claude", OutputPath: "/o"},
			true,
		},
		{
			"missing output",
			&WorkflowConfig{Category: "workflows", Preset: "all", Mode: "static"},
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestHookConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *HookConfig
		wantErr bool
	}{
		{
			"happy install",
			&HookConfig{Category: "hooks", Preset: "linting", Locale: "en", Install: InstallScopeProject},
			false,
		},
		{
			"happy preview",
			&HookConfig{Category: "hooks", Preset: "all", Locale: "en", OutputPath: "/o"},
			false,
		},
		{
			"missing category",
			&HookConfig{Preset: "linting", Locale: "en", Install: InstallScopeProject},
			true,
		},
		{
			"missing preset",
			&HookConfig{Category: "hooks", Locale: "en", Install: InstallScopeProject},
			true,
		},
		{
			"invalid preset",
			&HookConfig{Category: "hooks", Preset: "telemetry", Locale: "en", Install: InstallScopeProject},
			true,
		},
		{
			"invalid install scope",
			&HookConfig{Category: "hooks", Preset: "linting", Locale: "en", Install: "system"},
			true,
		},
		{
			"missing locale",
			&HookConfig{Category: "hooks", Preset: "linting", Install: InstallScopeProject},
			true,
		},
		{
			"unsupported locale",
			&HookConfig{Category: "hooks", Preset: "linting", Locale: "fr", Install: InstallScopeProject},
			true,
		},
		{
			"neither install nor output",
			&HookConfig{Category: "hooks", Preset: "linting", Locale: "en"},
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
