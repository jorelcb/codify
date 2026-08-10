// Decidir sobre las propuestas de cambio (FR-010, FR-012, FR-014, FR-015).
//
// Vive **fuera de la corriente** a propósito. La corriente es append-only y registra qué pasó;
// esto es lo que está esperando a una persona ahora mismo, y cambia con cada decisión.
// Mezclarlos obligaría a reescribir bloques ya emitidos, que es justo lo que la corriente no
// hace.
//
// El núcleo está **bloqueado** esperando cada una de estas decisiones: `Prompter::present`
// emitió el `proposal.new` y aguarda en un canal. Por eso decidir no es una preferencia que se
// guarda para después — es lo que desatasca el turno.

import { t } from "./i18n.js";

const { invoke } = window.__TAURI__.core;

const panel = document.getElementById("decide");
const countEl = document.getElementById("decide-count");
const targetEl = document.getElementById("decide-target");
const rationaleEl = document.getElementById("decide-rationale");
const diffEl = document.getElementById("decide-diff");
const editorEl = document.getElementById("decide-editor");
const textEl = document.getElementById("decide-text");
const prevEl = document.getElementById("decide-prev");
const nextEl = document.getElementById("decide-next");
const approveEl = document.getElementById("decide-approve");
const editEl = document.getElementById("decide-edit");
const rejectEl = document.getElementById("decide-reject");

/** Cola de propuestas sin decidir, en el orden en que llegaron. */
let queue = [];
let index = 0;
let onDecided = () => {};

export function configure(handler) {
  if (handler) onDecided = handler;
}

/** Cuántas esperan decisión. Es lo que el contador muestra y lo que bloquea el cierre. */
export function pendingCount() {
  return queue.length;
}

export function isOpen() {
  return !panel.hidden;
}

/** Encola una propuesta recién llegada del núcleo. */
export function enqueue(proposal) {
  queue.push(proposal);
  render();
}

/**
 * Pinta el estado actual del panel.
 *
 * Exportado porque este módulo escribe texto a mano (el contador, la ruta, el motivo) y
 * `i18n.apply()` no lo alcanza: sin repintar, cambiar de idioma dejaría media decisión en el
 * idioma anterior.
 */
export function render() {
  panel.hidden = queue.length === 0;
  if (!queue.length) {
    countEl.textContent = "";
    return;
  }

  if (index >= queue.length) index = queue.length - 1;
  if (index < 0) index = 0;
  const p = queue[index];

  // El contador dice cuántas quedan **y** en cuál estás: navegar no puede desorientar.
  countEl.textContent = `${index + 1}/${queue.length} ${t("proposal.pending")}`;
  targetEl.textContent = p.target;
  rationaleEl.textContent = `${t("proposal.rationale")}: ${p.rationale}`;
  diffEl.textContent = p.unified;

  // Navegar solo tiene sentido con más de una.
  prevEl.disabled = queue.length < 2;
  nextEl.disabled = queue.length < 2;

  // El editor arranca cerrado y con el texto vacío: si arrastrara lo escrito para otra
  // propuesta, se aplicaría a la equivocada.
  editorEl.hidden = true;
  textEl.value = "";
}

/** Manda la decisión al núcleo y saca la propuesta de la cola. */
async function send(verdict, edited) {
  const p = queue[index];
  if (!p) return;

  try {
    await invoke("decide", { proposalId: p.id, verdict, edited: edited ?? null });
  } catch (err) {
    // No se saca de la cola: si el núcleo no la recibió, sigue esperando.
    onDecided({ error: String(err), proposal: p });
    return;
  }

  queue.splice(index, 1);
  if (index >= queue.length) index = Math.max(0, queue.length - 1);
  render();
  onDecided({ verdict, proposal: p });
}

/** Navegar **no pierde** lo ya decidido: solo cambia cuál se está mirando (T043). */
function move(step) {
  if (queue.length < 2) return;
  index = (index + step + queue.length) % queue.length;
  render();
  diffEl.focus();
}

prevEl.addEventListener("click", () => move(-1));
nextEl.addEventListener("click", () => move(1));
approveEl.addEventListener("click", () => send("approve"));
rejectEl.addEventListener("click", () => send("reject"));

// Editar es en dos pasos: primero se abre el editor, luego se confirma. Enviar a la primera
// aplicaría un texto vacío, que el núcleo rechaza — mejor no llegar a pedirlo.
editEl.addEventListener("click", () => {
  if (editorEl.hidden) {
    editorEl.hidden = false;
    textEl.focus();
    return;
  }
  if (!textEl.value.trim()) {
    textEl.focus();
    return;
  }
  send("edit", textEl.value);
});

/** Vacía la cola al empezar una sesión nueva. */
export function reset() {
  queue = [];
  index = 0;
  render();
}
