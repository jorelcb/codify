package main

import (
	"os"

	"github.com/jorelcb/ai-context-generator/internal/interfaces/cli"
)

// Version information (will be set by build flags)
var (
	version = "dev"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	// Pass version info to CLI
	cli.Version = version
	cli.Commit = commit
	cli.Date = date

	// Execute CLI
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}