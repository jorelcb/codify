// Lo que se aplicó **sin preguntar** (FR-008).
//
// El loop auto-aplica lo de bajo riesgo para no interrumpir por cada nimiedad. El precio de esa
// comodidad es que el usuario no dijo que sí: se le aplicó algo sin consultarle. La
// compensación es que deshacerlo esté **a mano** — no escondido en un menú, y no solo
// anunciado en una etiqueta.

import { t } from "./i18n.js";

const { invoke } = window.__TAURI__.core;

const panel = document.getElementById("applied");
const list = document.getElementById("applied-list");

let sessionId = null;
let aplicadas = [];
let onReverted = () => {};

export function configure(id, handler) {
  sessionId = id;
  if (handler) onReverted = handler;
}

export function reset() {
  aplicadas = [];
  render();
}

/** Registra un cambio que entró sin preguntar. */
export function add(proposal) {
  aplicadas.push(proposal);
  render();
}

/**
 * Repinta la lista. Exportado porque el texto se escribe a mano y `i18n.apply()` no lo
 * alcanza: sin esto, cambiar de idioma dejaría los botones en el anterior.
 */
export function render() {
  panel.hidden = aplicadas.length === 0;
  list.replaceChildren(
    ...aplicadas.map((p) => {
      const li = document.createElement("li");

      const ruta = document.createElement("span");
      ruta.className = "applied-target";
      ruta.textContent = p.target;

      const boton = document.createElement("button");
      boton.className = "ghost";
      boton.textContent = t("proposal.revert");
      boton.addEventListener("click", () => undo(p));

      li.append(ruta, boton);
      return li;
    }),
  );
}

/** Deshace y saca de la lista. El núcleo decide si se puede; aquí solo se pide. */
async function undo(p) {
  if (!sessionId) return;
  try {
    await invoke("revert_proposal", { sessionId, proposalId: p.id });
  } catch (err) {
    // No se saca de la lista: si el núcleo no lo deshizo, sigue aplicado.
    onReverted({ error: String(err), proposal: p });
    return;
  }
  aplicadas = aplicadas.filter((x) => x.id !== p.id);
  render();
  onReverted({ proposal: p });
}
