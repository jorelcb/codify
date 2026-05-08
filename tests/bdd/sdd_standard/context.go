package sdd_standard

import (
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/sdd"
)

// FeatureContext mantiene el estado entre steps de un mismo escenario.
// Se reinicia antes de cada scenario via reset() para garantizar
// independencia (godog corre scenarios en orden no-determinístico).
type FeatureContext struct {
	// registry es el SDD adapter registry bajo prueba. Se inicializa con
	// NewDefaultRegistry para que matchee el wiring real de codify.
	registry *sdd.Registry

	// standard es el adapter resuelto en el último Lookup/Resolve.
	standard service.SpecStandard

	// resolveErr captura cualquier error producido por Resolve o Lookup
	// para que los Then steps puedan inspeccionarlo.
	resolveErr error
}

// SetupTest se llama una vez antes de TODOS los scenarios. Acá no
// inicializamos el registry porque queremos un registry fresh por scenario
// (algunos tests podrían registrar adapters custom y no deben filtrar
// estado al siguiente).
func (f *FeatureContext) SetupTest() {}

// reset borra el estado mutable del contexto antes de cada scenario.
func (f *FeatureContext) reset() {
	f.registry = nil
	f.standard = nil
	f.resolveErr = nil
}
