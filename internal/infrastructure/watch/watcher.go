// Package watch implementa un foreground file watcher que monitorea los
// paths registrados en .codify/state.json (input_signals + artifacts) y
// dispara un callback cuando detecta cambios.
//
// Diseño y rationale en docs/adr/0008-watch-model-decision.md.
//
// El watcher es deliberadamente simple:
//   - Foreground (sin daemon, sin PID file, sin detach)
//   - Scope acotado a paths conocidos (no recursive walk del repo)
//   - Debounce configurable (default 2s) para coalescer eventos
//   - Sale limpio en cancelación de context (Ctrl+C → context.Cancel)
package watch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Event es la abstracción que emite el watcher cuando dispara el callback.
// Path es el archivo afectado (relativo o absoluto, según se haya registrado).
// Triggered es el momento en que se dispara después del debounce.
type Event struct {
	Triggered time.Time
	Paths     []string // archivos que cambiaron en la ventana de debounce
}

// Options parametriza la creación de un Watcher.
type Options struct {
	// Paths es la lista de archivos a monitorear. El watcher resuelve el
	// directorio padre de cada uno y subscribe a esos directorios via fsnotify
	// (necesario porque fsnotify reporta events sobre directorios, no archivos
	// individuales en muchas plataformas).
	Paths []string

	// Debounce es la ventana de quiet-time antes de disparar el callback.
	// Default 2s si se pasa zero.
	Debounce time.Duration

	// OnEvent es el callback invocado por cada batch debounceado.
	// Se ejecuta en una goroutine dedicada — no bloquea el event loop.
	OnEvent func(Event)

	// OnError es el callback opcional para errores no-fatales del watcher
	// (e.g. fsnotify cierra un canal). Si nil, errores se ignoran.
	OnError func(error)
}

// Watcher encapsula el fsnotify watcher + el debounce loop.
type Watcher struct {
	fsw       *fsnotify.Watcher
	opts      Options
	watchSet  map[string]bool // paths que nos interesan (set lookup rápido)
	dirs      map[string]bool // dirs subscritos (uno por archivo único)
	pending   map[string]bool // paths cambiados desde el último flush
	flushAt   time.Time       // cuándo se debe flushear el batch
	debounce  time.Duration
}

// New crea un Watcher pero no lo inicia. Llamar Start() para arrancar el
// event loop.
func New(opts Options) (*Watcher, error) {
	if opts.OnEvent == nil {
		return nil, fmt.Errorf("watch: OnEvent callback is required")
	}
	if len(opts.Paths) == 0 {
		return nil, fmt.Errorf("watch: at least one path is required")
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("watch: fsnotify init: %w", err)
	}

	debounce := opts.Debounce
	if debounce <= 0 {
		debounce = 2 * time.Second
	}

	w := &Watcher{
		fsw:      fsw,
		opts:     opts,
		watchSet: make(map[string]bool),
		dirs:     make(map[string]bool),
		pending:  make(map[string]bool),
		debounce: debounce,
	}

	// Resolver paths a absolutos para comparación robusta
	for _, p := range opts.Paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		w.watchSet[abs] = true
		w.dirs[filepath.Dir(abs)] = true
	}

	// Subscribirse al directorio padre de cada path. Si el directorio no
	// existe, lo skipeamos silenciosamente (best-effort).
	for dir := range w.dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		if err := w.fsw.Add(dir); err != nil {
			// Best-effort: si un dir falla, seguimos con los demás.
			if w.opts.OnError != nil {
				w.opts.OnError(fmt.Errorf("watch: subscribe %s: %w", dir, err))
			}
		}
	}

	if len(w.fsw.WatchList()) == 0 {
		_ = w.fsw.Close()
		return nil, fmt.Errorf("watch: no watchable directories from %d input paths", len(opts.Paths))
	}

	return w, nil
}

// Start corre el event loop hasta que ctx se cancele. Bloquea el caller —
// típicamente se invoca en main goroutine después de instalar un signal
// handler para SIGINT/SIGTERM.
func (w *Watcher) Start(ctx context.Context) error {
	defer w.fsw.Close()

	ticker := time.NewTicker(w.debounce / 4) // chequea con un cuarto del debounce
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.flushPending(true)
			return nil

		case ev, ok := <-w.fsw.Events:
			if !ok {
				return fmt.Errorf("watch: fsnotify events channel closed")
			}
			abs, err := filepath.Abs(ev.Name)
			if err != nil {
				continue
			}
			if !w.watchSet[abs] {
				// Evento sobre un archivo del mismo dir que no nos interesa.
				continue
			}
			if !isRelevantOp(ev.Op) {
				continue
			}
			w.pending[abs] = true
			w.flushAt = time.Now().Add(w.debounce)

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return fmt.Errorf("watch: fsnotify errors channel closed")
			}
			if w.opts.OnError != nil {
				w.opts.OnError(err)
			}

		case <-ticker.C:
			if len(w.pending) > 0 && time.Now().After(w.flushAt) {
				w.flushPending(false)
			}
		}
	}
}

// flushPending dispara el callback con los paths acumulados y reinicia el batch.
// Si finalCall=true, se invoca incluso con el batch vacío para señalar shutdown
// limpio (no usado actualmente pero reservado por si el caller quiere hook de
// cierre).
func (w *Watcher) flushPending(finalCall bool) {
	if len(w.pending) == 0 && !finalCall {
		return
	}
	if len(w.pending) == 0 {
		return
	}
	paths := make([]string, 0, len(w.pending))
	for p := range w.pending {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	w.pending = map[string]bool{}
	w.opts.OnEvent(Event{Triggered: time.Now(), Paths: paths})
}

// isRelevantOp filtra los Op codes de fsnotify que nos interesan.
// CHMOD se ignora (cambios de permisos no son drift relevante).
func isRelevantOp(op fsnotify.Op) bool {
	return op&fsnotify.Write != 0 ||
		op&fsnotify.Create != 0 ||
		op&fsnotify.Remove != 0 ||
		op&fsnotify.Rename != 0
}
