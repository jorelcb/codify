package commands

import (
	"fmt"
	"os"

	statedomain "github.com/jorelcb/codify/internal/domain/state"
	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
	"github.com/jorelcb/codify/internal/infrastructure/snapshot"
	infrastate "github.com/jorelcb/codify/internal/infrastructure/state"
)

// codifyVersionForState devuelve la versión del binario para popular
// state.json. Permite override via env var (útil en tests). Default "dev".
func codifyVersionForState() string {
	if v := os.Getenv("CODIFY_VERSION_OVERRIDE"); v != "" {
		return v
	}
	return codifyVersion
}

// writeProjectSnapshot construye y persiste .codify/state.json a partir del
// FS actual. Best-effort — si algo falla (e.g. permisos), emite warning a
// stderr pero no aborta el comando que lo invoca. El argumento generatedBy
// identifica quién disparó el snapshot ("init", "generate", "analyze",
// "reset-state").
func writeProjectSnapshot(generatedBy, projectName, preset, language, locale, target, kind, outputPath string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: snapshot skipped (could not resolve cwd): %v\n", err)
		return
	}

	state, err := snapshot.Build(snapshot.BuildOptions{
		ProjectPath: cwd,
		OutputPath:  outputPath,
		Project: statedomain.ProjectInfo{
			Name:     projectName,
			Preset:   preset,
			Language: language,
			Locale:   locale,
			Target:   target,
			Kind:     kind,
		},
		GeneratedBy:   generatedBy,
		CodifyVersion: codifyVersionForState(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: snapshot build failed: %v\n", err)
		return
	}

	statePath, err := infraconfig.ProjectStatePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: snapshot path resolution failed: %v\n", err)
		return
	}

	if err := infrastate.NewRepository().Save(statePath, state); err != nil {
		fmt.Fprintf(os.Stderr, "warning: snapshot save failed: %v\n", err)
		return
	}

	fmt.Printf("\n✓ Snapshot stored at %s (run 'codify check' to validate later)\n", statePath)
}
