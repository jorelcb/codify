package commands

import (
	domain "github.com/jorelcb/codify/internal/domain/config"
	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
)

// loadEffectiveConfig wraps Repository.LoadEffective con un fallback silencioso:
// si la lectura de configs falla por cualquier razón, devuelve los built-in
// defaults para que el flujo principal no se rompa por un YAML mal escrito.
//
// La razón: estos commands deben seguir funcionando aún cuando el config
// está corrupto. El usuario puede correr `codify config edit` para arreglarlo.
func loadEffectiveConfig() domain.Config {
	cfg, err := infraconfig.NewRepository().LoadEffective()
	if err != nil {
		return domain.BuiltinDefaults()
	}
	return cfg
}

// applyConfigDefaults rellena un campo string con el valor del config si:
//   1) la flag NO fue seteada explícitamente, Y
//   2) el target actualmente está vacío
//
// Esto es la mecánica de precedencia "flag > config > built-in" aplicada
// punto-a-punto. NO reemplaza la lógica del prompt interactivo: si después
// de aplicar config el campo sigue vacío y estamos en TTY, el caller puede
// seguir promptando.
func applyConfigDefaults(target *string, configValue string, explicit bool) {
	if explicit {
		return
	}
	if *target != "" {
		// Hubo un default de flag (e.g. "clean-ddd" en --preset). Solo
		// reemplazamos si el config dice algo distinto. Esto es subtle:
		// el "default del flag" es de menor prioridad que el config,
		// pero solo se distingue cuando NO fue explícito (caso ya
		// excluido arriba). Reemplazamos siempre que config tenga valor.
		if configValue != "" {
			*target = configValue
		}
		return
	}
	*target = configValue
}
