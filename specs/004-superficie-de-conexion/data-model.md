# Phase 1 — Modelo

Esta feature no toca el dominio. Las entidades de aquí son de **presentación**: viven en la piel y
existen porque hay que poder verificarlas.

---

## Modo

El único hecho de esta superficie con consecuencias de privacidad, y el que hoy tiene tres dueños
(research D2).

| | |
|---|---|
| **Valores** | `local` \| `hybrid` |
| **Fuente única** | `AppState.mode`, en el backend |
| **Quién lo escribe** | `set_mode` (comando), a petición del usuario |
| **Quién lo lee** | `mode()` para pintar; `start_session` para armar el grafo |
| **Copias en la interfaz** | **Ninguna.** Ni `ui.local`, ni el `checked` de la casilla como estado propio |

**Superficies** — dos, y ninguna es dueña:

| Superficie | Dónde | Para qué |
|---|---|---|
| Insignia | `#mode`, barra superior | Responder «¿en qué modo estoy?» de un vistazo |
| Control | `#modo-local`, panel de conexión | Responder «quiero cambiarlo» |

**Invariante (FR-003a, SC-006).** Las dos se pintan en la misma llamada, desde el mismo valor. No
existe camino por el que una cambie y la otra no.

**Transición.** El usuario actúa sobre el control → `set_mode` → el backend devuelve el modo
resultante → se pintan las dos superficies. La sesión viva conserva su modo (`003`-FR-008b); el
nuevo rige la siguiente (`003`-FR-008a) — que es lo que hasta hoy no ocurría.

---

## Región

Una zona de la interfaz recorrible por un lector de pantalla.

| | |
|---|---|
| **Identidad** | Su **nombre accesible**, resuelto del catálogo vía `data-i18n-aria` |
| **Regla de unicidad** | Dos regiones no comparten nombre. Ni la clave, ni el texto, en ninguno de los dos idiomas |
| **Se comprueba** | Automáticamente (FR-006). Un nombre repetido hace fallar el build |

**Cómo se nombran, tras el ciclo.** Cada región se nombra **desde su propio encabezado** con
`aria-labelledby`. Así el nombre tiene un solo dueño y la región aparece en los **dos** rotores: el
de puntos de referencia y el de títulos. Nombrarlas con una clave suelta las dejaba fuera del
segundo, que es el que se usa para hacerse un mapa de una pantalla.

| Región | Encabezado que la nombra |
|---|---|
| `header.bar` | `a11y.toolbar_region` |
| `#provider` | `provider.title` |
| `#conexiones` | `a11y.connections_region` |
| `#stream` | `a11y.stream_region` |
| `#decide` | `proposal.title` |
| `#applied` | `proposal.applied_title` |
| `footer.prefs` | `a11y.prefs_region` |

Seis regiones visibles a la vez —`#decide` y `#applied` solo durante una sesión— y **el diálogo de
artefacto ya no es una**: llevaba `role="region"` y figuraba en el rotor estando cerrado.

Se retiraron `a11y.provider_region`, `a11y.decide_region`, `a11y.applied_region` y
`a11y.artifact_region`: nombraban regiones que **ya tenían encabezado propio**, así que eran una
segunda copia del mismo nombre. Una ya había divergido — decía «Configuración del modelo» donde la
vista dice «Modelo».

---

## Superficie de conexión

Hoy no existe como unidad: está fundida con el indicador de proveedor, del que solo la separa un
comentario en el HTML. Pasa a ser una zona con presentación propia y tres partes, en este orden:

1. **Modo** — arriba y separado. Va primero porque decide si algo puede salir del equipo, y verlo
   después de conectar sería tarde.
2. **Cuentas conectadas** — lista, o el texto de «ninguna» cuando está vacía.
3. **Formulario** — plegado. Se despliega al pedirlo (FR-002a).

**Estado por defecto**: sin cuentas y con el formulario plegado. Es el caso común, y es el que
debe producir menos ruido.
