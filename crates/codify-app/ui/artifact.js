// Vista de un artefacto completo (FR-021, FR-011 a FR-013).
//
// Es el contrapeso de la corriente cronológica. En un flujo que avanza, los artefactos quedan
// atrás; esta vista permite abrir uno **en cualquier momento** sin recorrerla hacia atrás.
//
// La regla que gobierna el render: **la etiqueta textual es la señal primaria**. El glifo y el
// color van encima, no en lugar de. Quitar el color —daltonismo, escala de grises, una captura
// impresa— no puede volver ambiguo si algo está verificado o no (FR-026, SC-002).
//
// Nada de lógica de dominio: qué es fundamentado, tentativo o contradictorio lo decidió el
// núcleo y viaja en `kind`. Aquí solo se elige cómo se ve.

import { t } from "./i18n.js";

const { invoke } = window.__TAURI__.core;

const dialog = document.getElementById("artifact-dialog");
const selectEl = document.getElementById("artifact-path");
const stateEl = document.getElementById("artifact-write-state");
const bodyEl = document.getElementById("artifact-body");
const emptyEl = document.getElementById("artifact-empty");

/** Estado de fundamento → clave de etiqueta + glifo. El color lo pone el CSS. */
const KINDS = {
  grounded: { label: "artifact.grounded", glyph: "✓" },
  tentative: { label: "artifact.tentative", glyph: "?" },
  contradiction: { label: "artifact.contradiction", glyph: "≠" },
};

/** Estado de escritura → clave. Un archivo en pantalla no es un archivo en el repositorio. */
const WRITE_STATES = {
  written: "artifact.in_repository",
  pending: "artifact.not_written",
  failed: "artifact.write_failed",
  skipped: "artifact.write_skipped",
};

let sessionId = null;
let current = null; // último artefacto cargado
let onChange = () => {}; // aviso a la orquestación cuando cambia el recuento

export function configure(id, handler) {
  sessionId = id;
  if (handler) onChange = handler;
}

/** ¿Hay artefactos que ofrecer? La orquestación lo usa para habilitar el control de apertura. */
export function setAvailable(paths) {
  selectEl.replaceChildren(
    ...paths.map((p) => {
      const opt = document.createElement("option");
      opt.value = p;
      opt.textContent = p;
      return opt;
    }),
  );
  selectEl.disabled = paths.length === 0;
  return paths.length;
}

function metaLine(labelKey, value) {
  const p = document.createElement("p");
  p.className = "meta";
  const label = document.createElement("span");
  label.className = "meta-label";
  label.textContent = `${t(labelKey)}: `;
  p.append(label, document.createTextNode(value));
  return p;
}

/**
 * Pinta un fragmento con sus **tres señales redundantes**: etiqueta, glifo y color.
 * La etiqueta va primero en el orden de lectura, que es lo que la hace primaria de verdad.
 */
function renderSegment(segment, index) {
  const spec = KINDS[segment.kind] ?? KINDS.tentative;
  const el = document.createElement("article");
  el.className = "segment";
  el.dataset.kind = segment.kind;
  el.tabIndex = 0; // recorrible con teclado (FR-025)

  const mark = document.createElement("p");
  mark.className = "mark";
  const glyph = document.createElement("span");
  glyph.className = "glyph";
  glyph.setAttribute("aria-hidden", "true"); // decorativo: la etiqueta ya lo dice
  glyph.textContent = spec.glyph;
  mark.append(glyph, document.createTextNode(t(spec.label)));

  // Un tentativo ya diferido sigue sin estar verificado: se dice que está pendiente
  // **a sabiendas**, no se le quita la marca.
  if (segment.kind === "tentative" && segment.acknowledged) {
    const extra = document.createElement("span");
    extra.className = "deferred";
    extra.textContent = ` · ${t("artifact.deferred")}`;
    mark.appendChild(extra);
  }
  el.appendChild(mark);

  const text = document.createElement("p");
  text.className = "text";
  text.textContent = segment.text;
  el.appendChild(text);

  // De qué fuente salió (FR-012) y, en una contradicción, qué fuentes chocan (FR-013).
  if (segment.sources?.length) {
    el.appendChild(metaLine("artifact.sources", segment.sources.join(", ")));
  }
  if (segment.reason) {
    el.appendChild(metaLine("artifact.reason", segment.reason));
  }

  if (segment.kind === "tentative" && !segment.acknowledged) {
    const action = document.createElement("button");
    action.className = "ghost defer";
    action.textContent = t("artifact.defer");
    action.title = t("artifact.defer_hint");
    action.addEventListener("click", () => defer(index));
    el.appendChild(action);
  }

  return el;
}

/** Difiere un fragmento y vuelve a cargar: el recuento lo manda el núcleo, no se estima aquí. */
async function defer(index) {
  if (!sessionId || !current) return;
  const remaining = await invoke("defer_tentative", {
    sessionId,
    path: current.path,
    index,
  });
  await load(current.path);
  onChange(remaining);
}

/**
 * Repinta lo ya cargado, **sin volver a pedirlo** al núcleo.
 *
 * Separado de `load` por el mismo motivo que en el panel del proveedor: las etiquetas de los
 * fragmentos y el estado de escritura se escriben a mano, así que `i18n.apply()` no los
 * alcanza. Sin esto, cambiar de idioma con la vista abierta dejaba el fundamento en el idioma
 * anterior — y el fundamento es justo lo que esta vista existe para comunicar.
 */
export function render() {
  if (!current) return;
  stateEl.textContent = t(WRITE_STATES[current.writeState] ?? "artifact.not_written");
  stateEl.dataset.state = current.writeState;

  const segments = current.segments ?? [];
  emptyEl.hidden = segments.length > 0;
  bodyEl.replaceChildren(...segments.map(renderSegment));
}

/** Carga un artefacto por su ruta y lo pinta. */
export async function load(path) {
  if (!sessionId || !path) return null;
  current = await invoke("artifact", { sessionId, path });
  render();
  return current;
}

/** Abre la vista. Alcanzable en cualquier momento mientras haya algo generado (FR-021). */
export async function open(path) {
  const target = path ?? selectEl.value;
  if (!target) return;
  selectEl.value = target;
  await load(target);
  dialog.showModal();
  // El foco entra en el diálogo: sin esto, la navegación por teclado se queda fuera.
  selectEl.focus();
}

export function close() {
  dialog.close();
}

export function isOpen() {
  return dialog.open;
}

/** Cuántos tentativos sin atender tiene el artefacto cargado. */
export function unattendedInView() {
  return (current?.segments ?? []).filter((s) => s.kind === "tentative" && !s.acknowledged)
    .length;
}

/** Lleva el foco al primer punto sin atender: «revisarlos» tiene que llevar a algún sitio. */
export function focusFirstUnattended() {
  const el = [...bodyEl.querySelectorAll('.segment[data-kind="tentative"]')].find(
    (s) => s.querySelector(".defer"),
  );
  el?.focus();
  el?.scrollIntoView({ block: "center", behavior: "smooth" });
}

selectEl.addEventListener("change", () => load(selectEl.value));
