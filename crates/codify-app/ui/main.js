// Scaffold de la piel (T003) + consumo de los comandos y eventos de T029.
// La UI diseñada llega en T030, guiada por specs/002-authoring-experience.
//
// La piel es tonta a propósito: renderiza lo que el núcleo emite y captura
// decisiones. Ninguna regla de dominio vive aquí.

const { invoke } = window.__TAURI__.core;
const { listen } = window.__TAURI__.event;

const stream = document.getElementById("stream");
const stateLabel = document.getElementById("state");
const startBtn = document.getElementById("start");
const repoInput = document.getElementById("repo");
const localCheck = document.getElementById("local");

/** Añade un bloque a la corriente cronológica. */
function block(kind, what, target, reason) {
  const el = document.createElement("article");
  el.className = "block";
  el.dataset.kind = kind;

  const head = document.createElement("div");
  head.className = "what";
  head.textContent = what;
  el.appendChild(head);

  if (target) {
    const t = document.createElement("div");
    t.className = "target";
    t.textContent = target;
    el.appendChild(t);
  }
  if (reason) {
    const r = document.createElement("div");
    r.className = "reason";
    r.textContent = reason;
    el.appendChild(r);
  }

  const hint = stream.querySelector(".hint");
  if (hint) hint.remove();
  stream.appendChild(el);
  el.scrollIntoView({ block: "end", behavior: "smooth" });
}

function setState(state) {
  stateLabel.textContent = state;
  stateLabel.dataset.state = state;
}

// --- Eventos del núcleo (contrato en 001/contracts/tauri-commands.md) ---

listen("agent.activity", (e) => {
  const { action, target } = e.payload;
  block("activity", action, target);
});

listen("reference.resolved", (e) => {
  block("resolved", "leído", e.payload.target);
});

listen("reference.unresolved", (e) => {
  const { target, reason } = e.payload;
  block("unresolved", "no resuelto", target, reason);
});

listen("egress.blocked", (e) => {
  block("egress", "salida bloqueada (modo local)", e.payload.target);
});

listen("session.state_changed", (e) => setState(e.payload.state));

// --- Comandos ---

startBtn.addEventListener("click", async () => {
  const repoRoot = repoInput.value.trim();
  if (!repoRoot) {
    block("error", "falta la ruta", "elige un repositorio antes de iniciar");
    return;
  }

  startBtn.disabled = true;
  setState("iniciando");

  try {
    const sessionId = await invoke("start_session", {
      request: { repoRoot, local: localCheck.checked },
    });

    const snapshot = await invoke("session_state", { sessionId });
    setState(snapshot.state);

    block(
      "activity",
      "sesión terminada",
      `${snapshot.artifacts.length} artefacto(s) · idioma ${snapshot.locale ?? "?"}`,
      snapshot.budgetExhausted ? "la exploración se acotó: hay material sin leer" : null,
    );

    for (const omitted of snapshot.omitted) {
      block("unresolved", "no leído", omitted);
    }
  } catch (err) {
    setState("error");
    block("error", "la sesión falló", String(err));
  } finally {
    startBtn.disabled = false;
  }
});
