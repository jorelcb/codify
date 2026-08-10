// La corriente cronológica de bloques (FR-020).
//
// Es **append-only**: espeja la naturaleza del log de auditoría del que nace. Los bloques no
// se reescriben ni se reordenan; lo que pasó, pasó.
//
// Cada tipo se distingue por **tres señales redundantes** — etiqueta textual, glifo y color
// (FR-026). La etiqueta es la primaria: quitar el color no puede volver la interfaz ambigua,
// y así sobrevive al daltonismo y a una captura en escala de grises.

import { t } from "./i18n.js";

const stream = document.getElementById("stream");

/** Tipo de bloque → clave de etiqueta + glifo. El color lo pone el CSS. */
const KINDS = {
  activity: { label: "stream.activity", glyph: "·" },
  read: { label: "stream.read", glyph: "✓" },
  unresolved: { label: "stream.unresolved", glyph: "?" },
  contradiction: { label: "stream.contradiction", glyph: "≠" },
  written: { label: "stream.written", glyph: "▸" },
  interview: { label: "stream.interview", glyph: "✎" },
  proposal: { label: "proposal.label", glyph: "±" },
  egress: { label: "stream.egress_blocked", glyph: "⊘" },
  cancelled: { label: "stream.cancelled", glyph: "■" },
  error: { label: "stream.error", glyph: "!" },
};

let seq = 0;

function clearHint() {
  stream.querySelector(".hint")?.remove();
}

/**
 * Añade un bloque al final de la corriente.
 * @param {string} kind  clave de KINDS
 * @param {string} target  lo concreto: una ruta, una URL, un nombre
 * @param {string} [detail]  motivo o explicación, cuando la hay
 * @param {string} [labelOverride]  texto ya resuelto, para casos puntuales
 */
export function push(kind, target, detail, labelOverride) {
  const spec = KINDS[kind] ?? KINDS.activity;
  clearHint();

  const block = document.createElement("article");
  block.className = "block";
  block.dataset.kind = kind;
  block.tabIndex = 0; // alcanzable con teclado (FR-025)
  block.style.setProperty("--seq", String(++seq));

  const head = document.createElement("p");
  head.className = "what";
  const glyph = document.createElement("span");
  glyph.className = "glyph";
  glyph.setAttribute("aria-hidden", "true");
  glyph.textContent = spec.glyph;
  head.append(glyph, document.createTextNode(labelOverride ?? t(spec.label)));
  block.appendChild(head);

  if (target) {
    const el = document.createElement("p");
    el.className = "target";
    el.textContent = target;
    block.appendChild(el);
  }

  if (detail) {
    const el = document.createElement("p");
    el.className = "detail";
    el.textContent = detail;
    block.appendChild(el);
  }

  stream.appendChild(block);
  block.scrollIntoView({ block: "end", behavior: "smooth" });
  return block;
}

/** Bloque compuesto: un título y una lista. Se usa para los balances. */
export function pushList(kind, title, items, hint) {
  const block = push(kind, null, hint, title);
  if (!items.length) return block;

  const list = document.createElement("ul");
  list.className = "items";
  for (const item of items) {
    const li = document.createElement("li");
    li.textContent = item;
    list.appendChild(li);
  }
  block.appendChild(list);
  return block;
}

/** Vacía la corriente al iniciar una sesión nueva. */
export function reset(hintKey = "stream.hint") {
  stream.replaceChildren();
  seq = 0;
  const hint = document.createElement("p");
  hint.className = "hint";
  hint.dataset.i18n = hintKey;
  hint.textContent = t(hintKey);
  stream.appendChild(hint);
}

/** Bloques navegables con teclado, en orden visual (FR-025). */
export function blocks() {
  return [...stream.querySelectorAll(".block")];
}
