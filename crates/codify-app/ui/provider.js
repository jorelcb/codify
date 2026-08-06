// Onboarding guiado del proveedor de modelo (FR-019).
//
// La regla que gobierna este módulo: **nunca quedarse en silencio**. Si no hay backend, el
// usuario tiene que ver qué pasa y qué hacer al respecto. Un usuario que abre la aplicación,
// no ve nada y no entiende por qué, no la vuelve a abrir.

import { t } from "./i18n.js";

const { invoke } = window.__TAURI__.core;

const statusEl = document.getElementById("provider-status");
const glyphEl = document.getElementById("provider-glyph");
const modelEl = document.getElementById("model");
const nextEl = document.getElementById("provider-next");
const panel = document.getElementById("provider");

let lastStatus = null;

/** ¿Hay backend listo para trabajar? Lo consulta la orquestación antes de iniciar. */
export function isReady() {
  return Boolean(lastStatus?.reachable && modelEl.value);
}

export function selectedModel() {
  return modelEl.value || null;
}

/**
 * Pinta el panel a partir del **último sondeo**, sin volver a sondear.
 *
 * Está separado de `refresh` porque este texto se escribe a mano y por tanto `i18n.apply()`
 * no lo alcanza: al cambiar de idioma se quedaba en el anterior. Peor todavía, el `data-i18n`
 * que llevaba el elemento lo devolvía a «comprobando…» mientras el glifo seguía en ✓ — el
 * panel afirmaba dos cosas distintas a la vez.
 */
export function render() {
  if (!lastStatus) {
    panel.dataset.state = "checking";
    glyphEl.textContent = "…";
    statusEl.textContent = t("provider.checking");
    nextEl.hidden = true;
    return;
  }

  if (lastStatus.reachable && lastStatus.models.length) {
    panel.dataset.state = "ready";
    glyphEl.textContent = "✓";
    statusEl.textContent = `${t("provider.reachable")} · ${lastStatus.endpoint}`;
    nextEl.hidden = true;
    return;
  }

  // Sin backend: el motivo es obligatorio y se muestra como algo accionable.
  panel.dataset.state = "down";
  glyphEl.textContent = "!";
  statusEl.textContent = lastStatus.models.length
    ? t("provider.no_models")
    : `${t("provider.unreachable")} · ${lastStatus.endpoint}`;
  // El núcleo nombra el motivo; el texto lo elige la piel, así que cambia con el idioma.
  const queHacer = lastStatus.issue ? t(`provider.issue.${lastStatus.issue}`) : t("error.unknown");
  nextEl.textContent = `${t("provider.next_step")}: ${queHacer}`;
  nextEl.hidden = false;
}

/** Sondea el backend y refleja el resultado. No lanza: informar es su trabajo. */
export async function refresh(local = true) {
  lastStatus = null;
  render(); // «comprobando…», para que el sondeo no transcurra en silencio

  let status;
  try {
    status = await invoke("probe_provider", { local });
  } catch (err) {
    status = { reachable: false, endpoint: "?", models: [], detail: String(err) };
  }
  lastStatus = status;

  // La lista de modelos solo se repuebla al sondear: hacerlo al repintar borraría la
  // elección del usuario cada vez que cambia el idioma.
  const usable = status.reachable && status.models.length;
  modelEl.replaceChildren(
    ...(usable
      ? status.models.map((m) => {
          const opt = document.createElement("option");
          opt.value = m;
          opt.textContent = m;
          return opt;
        })
      : []),
  );
  modelEl.disabled = !usable;

  render();
  return status;
}
