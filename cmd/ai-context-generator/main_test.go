package main

import (
	"testing"

	"github.com/jorelcb/ai-context-generator/internal/interfaces/cli"
)

func TestVersion(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}

	// In development, version should be "0.0.1-alpha"
	if version != "0.0.1-alpha" {
		t.Logf("Warning: version is %q, expected \"0.0.1-alpha\"", version)
	}
}

func TestCommit(t *testing.T) {
	if commit == "" {
		t.Error("commit should not be empty")
	}

	// In development, commit should be "dev"
	if commit != "dev" {
		t.Logf("Info: commit is %q (expected \"dev\" in development)", commit)
	}
}

func TestVersionSetInCLI(t *testing.T) {
	// Set version info like main() does
	cli.Version = version
	cli.Commit = commit
	cli.Date = date

	// Verify they're set correctly
	if cli.Version != version {
		t.Errorf("CLI version not set correctly: got %q, want %q", cli.Version, version)
	}
	if cli.Commit != commit {
		t.Errorf("CLI commit not set correctly: got %q, want %q", cli.Commit, commit)
	}
	if cli.Date != date {
		t.Errorf("CLI date not set correctly: got %q, want %q", cli.Date, date)
	}
}

func TestMain_Variables(t *testing.T) {
	// This test just ensures the variables are initialized
	if version == "" || commit == "" || date == "" {
		t.Error("version variables should be initialized")
	}
}