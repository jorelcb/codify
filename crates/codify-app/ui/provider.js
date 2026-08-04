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

/** Sondea el backend y refleja el resultado. No lanza: informar es su trabajo. */
export async function refresh(local = true) {
  panel.dataset.state = "checking";
  glyphEl.textContent = "…";
  statusEl.textContent = t("provider.checking");
  nextEl.hidden = true;

  let status;
  try {
    status = await invoke("probe_provider", { local });
  } catch (err) {
    status = { reachable: false, endpoint: "?", models: [], detail: String(err) };
  }
  lastStatus = status;

  if (status.reachable && status.models.length) {
    panel.dataset.state = "ready";
    glyphEl.textContent = "✓";
    statusEl.textContent = `${t("provider.reachable")} · ${status.endpoint}`;
    modelEl.replaceChildren(
      ...status.models.map((m) => {
        const opt = document.createElement("option");
        opt.value = m;
        opt.textContent = m;
        return opt;
      }),
    );
    modelEl.disabled = false;
    return status;
  }

  // Sin backend: el motivo es obligatorio y se muestra como algo accionable.
  panel.dataset.state = "down";
  glyphEl.textContent = "!";
  statusEl.textContent = status.models.length
    ? t("provider.no_models")
    : `${t("provider.unreachable")} · ${status.endpoint}`;
  modelEl.replaceChildren();
  modelEl.disabled = true;

  nextEl.textContent = `${t("provider.next_step")}: ${status.detail ?? t("error.unknown")}`;
  nextEl.hidden = false;
  return status;
}
