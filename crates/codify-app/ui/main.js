// Orquestación de la piel.
//
// La piel es **tonta a propósito**: renderiza lo que el núcleo emite y captura decisiones.
// Ninguna regla de dominio vive aquí — qué está fundamentado, qué es de alto impacto y qué se
// escribe lo decide el núcleo (constitución, Principio I).

import * as i18n from "./i18n.js";
import * as stream from "./stream.js";
import * as provider from "./provider.js";

const { invoke } = window.__TAURI__.core;
const { listen } = window.__TAURI__.event;
const appWindow = window.__TAURI__.window?.getCurrentWindow?.();

const el = {
  repo: document.getElementById("repo"),
  action: document.getElementById("action"),
  state: document.getElementById("state"),
  mode: document.getElementById("mode"),
  modeLabel: document.getElementById("mode-label"),
  uiLocale: document.getElementById("ui-locale"),
  artifactLocale: document.getElementById("artifact-locale"),
  providerRetry: document.getElementById("provider-retry"),
  closeDialog: document.getElementById("close-dialog"),
  closeConfirm: document.getElementById("close-confirm"),
  closeCancel: document.getElementById("close-cancel"),
};

/** Estado de la piel. El estado real de la sesión vive en el núcleo. */
const ui = {
  sessionId: null,
  running: false,
  local: true,
};

// ---------------------------------------------------------------------------
// Presentación del estado
// ---------------------------------------------------------------------------

/**
 * Presenta el estado de la sesión.
 *
 * `idle` no es una fase del núcleo sino la **ausencia** de sesión, así que tiene su propia
 * cadena. Resolverla aquí y no en el arranque es lo que hace que sobreviva a un cambio de
 * idioma: si la asignara `boot()` a mano, se quedaría congelada en el idioma inicial.
 */
function setSessionState(state) {
  el.state.dataset.state = state;
  el.state.textContent = i18n.t(state === "idle" ? "session.none" : `session.state.${state}`);
}

/** El indicador de modo es **persistente** y lleva texto, no solo color (FR-005/FR-026). */
function renderMode() {
  const key = ui.local ? "mode.local" : "mode.hybrid";
  el.mode.dataset.mode = ui.local ? "local" : "hybrid";
  el.modeLabel.textContent = i18n.t(key);
  el.mode.title = i18n.t(ui.local ? "mode.local_hint" : "mode.hybrid_hint");
}

function setRunning(running) {
  ui.running = running;
  el.action.textContent = i18n.t(running ? "session.cancel" : "session.start");
  el.action.dataset.role = running ? "cancel" : "start";
  el.repo.disabled = running;
}

/** Ningún fallo se muestra crudo: se traduce a qué pasó y qué hacer (FR-028). */
function showError(messageKey, detail) {
  stream.push("error", i18n.t(messageKey), detail);
}

// ---------------------------------------------------------------------------
// Balances: lo escrito y lo no leído
// ---------------------------------------------------------------------------

/** Qué llegó al repositorio (FR-017 / FR-023). */
function renderBalance(writes) {
  if (!writes || !writes.length) {
    stream.pushList("written", i18n.t("session.balance.title"), [], i18n.t("session.balance.none"));
    return;
  }
  const items = writes.map((w) => {
    if (w.outcome === "written") return `${w.path} — ${i18n.t("session.balance.written")} (${w.bytes} B)`;
    const verb = w.outcome === "skipped" ? "session.balance.skipped" : "session.balance.failed";
    return `${w.path} — ${i18n.t(verb)}: ${w.detail ?? ""}`;
  });
  stream.pushList("written", i18n.t("session.balance.title"), items);
}

/** Lo que quedó sin leer: el resultado nunca debe aparentar ser completo (FR-004). */
function renderOmitted(snapshot) {
  if (!snapshot.omitted?.length && !snapshot.budgetExhausted) return;
  stream.pushList(
    "unresolved",
    i18n.t("stream.omitted.title"),
    snapshot.omitted ?? [],
    i18n.t("stream.omitted.hint"),
  );
}

// ---------------------------------------------------------------------------
// Eventos del núcleo
// ---------------------------------------------------------------------------

listen("agent.activity", (e) => stream.push("activity", e.payload.target, null, e.payload.action));
listen("reference.resolved", (e) => stream.push("read", e.payload.target));
listen("reference.unresolved", (e) => stream.push("unresolved", e.payload.target, e.payload.reason));
listen("artifact.written", (e) => stream.push("written", e.payload.target, e.payload.reason));
listen("session.cancelled", (e) => stream.push("cancelled", null, e.payload.target));
listen("session.state_changed", (e) => setSessionState(e.payload.state));

/** El intento de salida bloqueado se ve: es la garantía de cero-egress hecha visible (SC-006). */
listen("egress.blocked", (e) => stream.push("egress", e.payload.target, i18n.t("mode.local_hint")));

// ---------------------------------------------------------------------------
// Acciones
// ---------------------------------------------------------------------------

async function start() {
  const repoRoot = el.repo.value.trim();
  if (!repoRoot) {
    showError("error.no_repo");
    el.repo.focus();
    return;
  }
  if (!provider.isReady()) {
    await provider.refresh(ui.local);
    if (!provider.isReady()) return; // el panel ya explica qué falta (FR-019)
  }

  stream.reset("session.state.ingesting");
  setRunning(true);
  setSessionState("ingesting");

  try {
    ui.sessionId = await invoke("start_session", {
      request: { repoRoot, local: ui.local, locale: el.artifactLocale.value },
    });
    await finish();
  } catch (err) {
    setRunning(false);
    setSessionState("failed");
    showError("error.session_failed", String(err));
  }
}

/** Cancelar es alcanzable **en cualquier momento** mientras hay trabajo (FR-023). */
async function cancel() {
  if (!ui.sessionId) return;
  el.action.disabled = true;
  el.action.textContent = i18n.t("session.cancelling");
  try {
    const outcome = await invoke("cancel_session", { sessionId: ui.sessionId });
    setSessionState("cancelled");
    renderBalance(outcome.writes);
  } catch (err) {
    showError("error.unknown", String(err));
  } finally {
    el.action.disabled = false;
    setRunning(false);
  }
}

/** Consulta el resultado y presenta los balances. */
async function finish() {
  const snapshot = await invoke("session_state", { sessionId: ui.sessionId });
  setSessionState(snapshot.state);

  // Un repositorio vacío no es un error ni un logro: es que no hay material. Presentarlo con
  // el mismo gris que la actividad rutinaria lo volvía indistinguible del ruido, y el estado
  // «terminada» sugería un resultado que no existe (FR-004). Lleva su propio tipo de bloque
  // y su propio estado, más el qué hacer (FR-019/FR-028).
  if (snapshot.interviewMode) {
    setSessionState("interview");
    stream.push("interview", i18n.t("session.interview"), i18n.t("session.interview_next"));
  }
  renderOmitted(snapshot);
  renderBalance(snapshot.writes);

  if (snapshot.unattendedTentative > 0) {
    stream.push("unresolved", null, i18n.t("artifact.pending_tentative"));
  }
  setRunning(false);
}

// ---------------------------------------------------------------------------
// Teclado: toda acción alcanzable sin ratón (FR-025)
// ---------------------------------------------------------------------------

el.action.addEventListener("click", () => (ui.running ? cancel() : start()));
el.providerRetry.addEventListener("click", () => provider.refresh(ui.local));
el.repo.addEventListener("keydown", (e) => {
  if (e.key === "Enter" && !ui.running) start();
});

document.addEventListener("keydown", (e) => {
  if (e.key === "Escape" && ui.running) {
    cancel();
    return;
  }
  // Recorrer la corriente con las flechas, sin ratón.
  if (e.key === "ArrowDown" || e.key === "ArrowUp") {
    const blocks = stream.blocks();
    if (!blocks.length) return;
    const i = blocks.indexOf(document.activeElement);
    const next = e.key === "ArrowDown" ? Math.min(i + 1, blocks.length - 1) : Math.max(i - 1, 0);
    blocks[i === -1 ? blocks.length - 1 : next]?.focus();
    e.preventDefault();
  }
});

// ---------------------------------------------------------------------------
// Idioma: el de la interfaz y el de los artefactos son independientes (FR-016/016b)
// ---------------------------------------------------------------------------

el.uiLocale.addEventListener("change", async () => {
  await i18n.setLocale(el.uiLocale.value);
  renderMode();
  setRunning(ui.running);
  // Sin condición: el estado se repinta **siempre**, incluido `idle`. Excluirlo dejaba
  // «sin sesión» congelado en el idioma de arranque (SC-009).
  setSessionState(el.state.dataset.state);
  // Todo lo que se escribe a mano hay que repintarlo: `apply()` solo alcanza al DOM marcado.
  provider.render();
});

el.artifactLocale.addEventListener("change", async () => {
  // Cambia el idioma del contexto generado, no el de la aplicación.
  if (ui.sessionId) {
    await invoke("set_locale", { sessionId: ui.sessionId, locale: el.artifactLocale.value });
  }
});

// ---------------------------------------------------------------------------
// Cierre con trabajo en curso: declarar qué se pierde (FR-024)
// ---------------------------------------------------------------------------

if (appWindow?.onCloseRequested) {
  appWindow.onCloseRequested(async (event) => {
    if (!ui.running) return;
    event.preventDefault();
    el.closeDialog.showModal();
  });
}

el.closeCancel?.addEventListener("click", () => el.closeDialog.close());
el.closeConfirm?.addEventListener("click", async () => {
  el.closeDialog.close();
  if (ui.sessionId) await invoke("cancel_session", { sessionId: ui.sessionId }).catch(() => {});
  appWindow?.destroy?.();
});

// ---------------------------------------------------------------------------
// Arranque
// ---------------------------------------------------------------------------

(async function boot() {
  const detected = await i18n.detectLocale();
  await i18n.setLocale(detected);
  el.uiLocale.value = i18n.locale();
  el.artifactLocale.value = i18n.locale();

  renderMode();
  setRunning(false);
  setSessionState("idle");

  // Sondear al arrancar: si falta el backend, el usuario lo sabe antes de intentar nada.
  await provider.refresh(ui.local);
  el.repo.focus();
})();
