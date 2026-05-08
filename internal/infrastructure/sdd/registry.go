package sdd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jorelcb/codify/internal/domain/service"
)

// DefaultStandardID identifies the standard used when no flag, no project
// config, and no user config select one. ADR-0011 fija OpenSpec como default
// para preservar el comportamiento de v1.x.
const DefaultStandardID = "openspec"

// Registry guarda los SpecStandard adapters disponibles, indexados por ID.
// Las entradas se cablean en NewDefaultRegistry; agregar un standard nuevo
// (Spec-Kit en C.3, custom Konfio en el futuro) solo requiere registrarlo
// allí.
type Registry struct {
	adapters map[string]service.SpecStandard
}

// NewDefaultRegistry construye el registry con los adapters que vienen con
// el binario. Adapters se cablean explícitamente acá; agregar uno nuevo
// (Konfio interno futuro, otros estándares) solo requiere editar esta
// función.
func NewDefaultRegistry() *Registry {
	r := &Registry{adapters: make(map[string]service.SpecStandard)}
	r.Register(NewOpenSpecAdapter())
	r.Register(NewSpecKitAdapter())
	return r
}

// Register agrega un adapter al registry. Los IDs duplicados sobrescriben
// silenciosamente — pensado para que un wiring custom (un cmd alternativo,
// un test, o el package manager de Track D) pueda override el default sin
// errors. Si el caller quiere semántica strict, debe verificar con Lookup
// antes de registrar.
func (r *Registry) Register(s service.SpecStandard) {
	r.adapters[s.ID()] = s
}

// Lookup retorna el adapter con el ID dado, o un error explícito que lista
// los IDs disponibles. El error ayuda al usuario a entender qué standards
// están registrados sin tener que mirar el código.
func (r *Registry) Lookup(id string) (service.SpecStandard, error) {
	if s, ok := r.adapters[id]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("unknown SDD standard %q (available: %s)", id, r.availableIDs())
}

// availableIDs devuelve los IDs registrados como una lista comma-separada
// en orden alfabético estable. Útil para mensajes de error.
func (r *Registry) availableIDs() string {
	return strings.Join(r.AvailableIDs(), ", ")
}

// AvailableIDs returns the list of registered IDs in alphabetical order.
// Exposed for callers that want the raw list (e.g., to render in CLI flag
// hints) instead of the comma-joined string used internally for errors.
func (r *Registry) AvailableIDs() []string {
	ids := make([]string, 0, len(r.adapters))
	for id := range r.adapters {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Resolve aplica la cadena de precedencia documentada en ADR-0011 y devuelve
// el SpecStandard activo:
//
//  1. flagValue (CLI flag --sdd-standard) si no está vacío
//  2. projectStandardID (.codify/config.yml > sdd_standard) si no está vacío
//  3. userStandardID (~/.codify/config.yml > sdd_standard) si no está vacío
//  4. DefaultStandardID (openspec)
//
// Cualquiera de los IDs intermedios que no exista en el registry produce un
// error claro — preferimos fallar rápido sobre fallback silencioso a un
// standard que el usuario no eligió.
func (r *Registry) Resolve(flagValue, projectStandardID, userStandardID string) (service.SpecStandard, error) {
	for _, candidate := range []string{flagValue, projectStandardID, userStandardID, DefaultStandardID} {
		if candidate == "" {
			continue
		}
		s, err := r.Lookup(candidate)
		if err != nil {
			return nil, fmt.Errorf("resolve SDD standard: %w", err)
		}
		return s, nil
	}
	// Inalcanzable — DefaultStandardID siempre matchea, pero defensivo
	// para que un futuro cambio de DefaultStandardID a "" no se vuelva
	// fallback silencioso.
	return nil, fmt.Errorf("no SDD standard could be resolved")
}
