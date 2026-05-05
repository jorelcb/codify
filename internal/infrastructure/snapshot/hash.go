// Package snapshot construye snapshots de estado del proyecto para
// .codify/state.json. Son funciones puras y determinísticas: para mismo FS
// y misma config devuelven mismo State.
//
// Los snapshots se consumen por los lifecycle commands (`check`, `update`,
// `audit`, `watch`) a partir de v1.23. La ruta es siempre:
//
//   1. generate/analyze produce artefactos
//   2. snapshot.Build() captura su estado al momento exacto
//   3. state repository persiste a .codify/state.json
//   4. check compara snapshot vs estado actual del FS
package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

// HashFile devuelve el SHA256 hex del contenido de path. Si el archivo no
// existe devuelve ("", false, nil) — el caller decide si la ausencia es
// válida (e.g. un input signal opcional). Errores de I/O se propagan.
func HashFile(path string) (string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", false, err
	}
	return hex.EncodeToString(h.Sum(nil)), true, nil
}

// FileSize devuelve el tamaño del archivo en bytes. Si no existe, (0, false, nil).
func FileSize(path string) (int64, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return info.Size(), true, nil
}

// CountLines devuelve un conteo aproximado de líneas en path. Útil como
// signal aproximado para README.md. Si no existe, (0, false, nil).
func CountLines(path string) (int, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, false, nil
		}
		return 0, false, err
	}
	count := 0
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	// Si el archivo no termina con \n pero tiene contenido, contar la última línea
	if len(data) > 0 && data[len(data)-1] != '\n' {
		count++
	}
	return count, true, nil
}
