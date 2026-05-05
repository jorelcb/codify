package snapshot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	statedomain "github.com/jorelcb/codify/internal/domain/state"
)

// BuildOptions parametriza la captura del snapshot. Todos los campos son
// opcionales excepto ProjectPath y OutputPath.
type BuildOptions struct {
	ProjectPath   string                  // raíz del proyecto (cwd típicamente)
	OutputPath    string                  // donde generate/analyze escribieron los artefactos
	Project       statedomain.ProjectInfo // metadata del proyecto (de config)
	GeneratedBy   string                  // "init", "generate", "analyze", "reset-state"
	CodifyVersion string                  // versión del binario
}

// Build captura un State completo desde el FS aplicando las opciones.
// Pura — no toca ningún archivo, solo lee.
//
// Errores en archivos individuales (e.g. permisos) se logean implícitamente
// vía el side-effect de no incluir esa entrada — el snapshot resultante es
// "best effort". Esto evita que un archivo con permiso denegado rompa el
// flujo entero. Errores reales (path inválido, etc.) sí se propagan.
func Build(opts BuildOptions) (statedomain.State, error) {
	state := statedomain.New()
	state.CodifyVersion = opts.CodifyVersion
	state.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	state.GeneratedBy = opts.GeneratedBy
	state.Project = opts.Project

	state.Git = captureGitInfo(opts.ProjectPath)

	artifacts, err := captureArtifacts(opts.OutputPath, opts.Project.Name)
	if err != nil {
		return statedomain.State{}, err
	}
	state.Artifacts = artifacts

	signals, err := captureInputSignals(opts.ProjectPath)
	if err != nil {
		return statedomain.State{}, err
	}
	state.InputSignals = signals

	return state, nil
}

// artifactGlobs lista los patrones que reconocemos como "artefactos generados
// por Codify". Capturarlos por nombre evita que el snapshot incluya archivos
// totalmente externos al control de Codify (e.g. el README del usuario, sus
// archivos de tests, etc).
//
// Cada entrada se evalúa contra outputPath/<projectName>/ (cuando aplica) y
// outputPath/ directamente.
var artifactGlobs = []string{
	"AGENTS.md",
	"context/CONTEXT.md",
	"context/INTERACTIONS_LOG.md",
	"context/DEVELOPMENT_GUIDE.md",
	"context/IDIOMS.md",
}

// captureArtifacts hashea cada archivo conocido bajo outputPath/{projectName}/
// y outputPath/. Solo incluye los que existen — los que no, no aparecen.
func captureArtifacts(outputPath, projectName string) (map[string]statedomain.ArtifactInfo, error) {
	out := make(map[string]statedomain.ArtifactInfo)
	roots := []string{}
	if projectName != "" {
		roots = append(roots, filepath.Join(outputPath, projectName))
	}
	roots = append(roots, outputPath)

	seen := map[string]bool{}
	for _, root := range roots {
		for _, glob := range artifactGlobs {
			path := filepath.Join(root, glob)
			if seen[path] {
				continue
			}
			seen[path] = true

			hash, exists, err := HashFile(path)
			if err != nil {
				continue // best-effort
			}
			if !exists {
				continue
			}
			size, _, _ := FileSize(path)

			// Key relativa al output root (sin el projectName) para que sea
			// consistente entre ejecuciones — incluso si el directorio
			// cambia de ubicación.
			key := strings.TrimPrefix(path, root+string(filepath.Separator))
			out[key] = statedomain.ArtifactInfo{
				SHA256:      hash,
				GeneratedAt: time.Now().UTC().Format(time.RFC3339),
				SizeBytes:   size,
			}
		}
	}
	return out, nil
}

// inputSignalFiles enumera los archivos cuyo cambio puede invalidar la
// validez del contexto generado. Cuando uno cambia, el contexto puede haber
// quedado desfasado y `codify check` lo reporta.
//
// La lista es modesta a propósito: archivos universales en la raíz del
// proyecto. v1.24+ puede agregar detección dinámica vía ProjectScanner
// (e.g. archivos.proto, schemas, etc.) si el feedback lo justifica.
var inputSignalFiles = []string{
	"go.mod",
	"go.sum",
	"package.json",
	"requirements.txt",
	"pyproject.toml",
	"Cargo.toml",
	"Gemfile",
	"pom.xml",
	"build.gradle",
	"composer.json",
	"Makefile",
	"Taskfile.yml",
	"Taskfile.yaml",
	"README.md",
}

// captureInputSignals hashea cada signal file en projectPath. Solo incluye
// los que existen.
func captureInputSignals(projectPath string) (map[string]statedomain.SignalInfo, error) {
	out := make(map[string]statedomain.SignalInfo)
	for _, name := range inputSignalFiles {
		path := filepath.Join(projectPath, name)
		hash, exists, err := HashFile(path)
		if err != nil || !exists {
			continue
		}
		info := statedomain.SignalInfo{SHA256: hash}
		// Métricas auxiliares por tipo de archivo
		switch name {
		case "README.md":
			if lines, ok, _ := CountLines(path); ok {
				info.Lines = lines
			}
		}
		out[name] = info
	}
	return out, nil
}

// captureGitInfo captura el commit, branch, remote, y dirty status del repo
// si está disponible. Si no es un repo git, devuelve un GitInfo zero.
func captureGitInfo(projectPath string) statedomain.GitInfo {
	if !isGitRepo(projectPath) {
		return statedomain.GitInfo{}
	}
	return statedomain.GitInfo{
		Commit:  runGit(projectPath, "rev-parse", "HEAD"),
		Branch:  runGit(projectPath, "rev-parse", "--abbrev-ref", "HEAD"),
		Remote:  runGit(projectPath, "remote", "get-url", "origin"),
		IsDirty: gitDirty(projectPath),
	}
}

// isGitRepo verifica si projectPath es repo git via la presencia de .git/.
func isGitRepo(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, ".git"))
	return err == nil
}

// runGit ejecuta un comando git y devuelve el output stripeado o "" en
// error. Best-effort: errores se silencian para que el snapshot funcione
// aún si git no está disponible.
func runGit(projectPath string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// gitDirty reporta si hay cambios sin commitear (working tree o index).
func gitDirty(projectPath string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}
