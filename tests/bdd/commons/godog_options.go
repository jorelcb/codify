package commons

import (
	"time"

	"github.com/cucumber/godog"
)

// Options returns standard godog options for test execution
func Options(paths ...string) *godog.Options {
	return &godog.Options{
		Format:    "pretty",                     // Human-readable output
		Paths:     paths,                        // Feature file paths
		Randomize: time.Now().UTC().UnixNano(), // Randomize scenario execution order
		Strict:    true,                         // Fail on undefined steps
	}
}