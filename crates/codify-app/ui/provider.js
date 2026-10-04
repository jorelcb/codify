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

// ---------------------------------------------------------------------------
// `006` — lo que hay en este equipo
// ---------------------------------------------------------------------------

/**
 * Pinta lo detectado **para elegir**, no para imponer.
 *
 * Encontrar algo no lo da por aceptado: los dos caminos —modelos que responden aquí y servicios
 * que consultan fuera— existen a propósito, y cuál se usa lo decide el usuario. Lo que sí se dice
 * siempre es **de dónde salió cada fila** y **dónde infiere**: una lista que aparece sola, sin
 * explicar nada, es la clase de afirmación que este producto existe para no hacer.
 */
export function renderDetectados(deteccion) {
  const lista = document.getElementById("detectados-lista");
  const estado = document.getElementById("detectados-estado");
  if (!lista || !estado) return;

  lista.replaceChildren(
    ...deteccion.modelos.map((m) => {
      const li = document.createElement("li");
      li.dataset.distancia = m.distancia;
      // Lo que solo **declara** su clasificación no se ofrece como local, aunque diga inferir
      // dentro. Donde se puede tener una garantía sostenida no se acepta una declarada.
      li.dataset.localSostenido = String(m.localSostenido);

      const nombre = document.createElement("span");
      nombre.className = "label";
      nombre.textContent = m.nombre;

      const detalle = document.createElement("span");
      detalle.className = "hint-inline";
      detalle.textContent = `${t(`detection.distance.${m.distancia}`)} · ${t(
        `detection.origin.${m.procedencia}`,
      )} · ${t(`detection.inference.${m.dondeInfiere}`)}`;

      li.append(nombre, detalle);

      // Lo que todavía no se puede usar dice **qué falta**: una lista de cosas inservibles sin
      // explicar qué les falta es tan inútil como no tenerla.
      if (m.distancia !== "usable") {
        const salida = document.createElement("p");
        salida.className = "next-step";
        salida.textContent = t(`detection.distance.${m.distancia}.next`);
        li.append(salida);
      }
      return li;
    }),
  );

  if (deteccion.incompleta) estado.textContent = t("detection.incomplete");
  else if (deteccion.modelos.length === 0) estado.textContent = t("detection.none");
  else estado.hidden = true;
}

/** Arranca la búsqueda sola: nadie tiene que pedirla. */
export async function detectar() {
  const estado = document.getElementById("detectados-estado");
  if (estado) {
    estado.hidden = false;
    estado.textContent = t("detection.searching");
  }
  try {
    renderDetectados(await invoke("detect_models"));
  } catch {
    renderDetectados({ modelos: [], incompleta: true });
  }
}
