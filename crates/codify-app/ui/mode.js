// `004` — el modo, y su única fuente.
//
// El modo se enseña en dos sitios porque responden preguntas distintas: la insignia de la barra
// dice en qué modo estás de un vistazo, la casilla del panel permite cambiarlo. Lo que no puede
// ocurrir es que digan cosas distintas.
//
// Antes ocurría. La insignia se pintaba desde una copia que la interfaz fijaba a `true` y no
// reasignaba jamás, así que decía «local» hicieras lo que hicieras; y `start_session` decidía el
// modo de la sesión leyendo esa misma copia, de modo que `set_mode` guardaba un valor que ningún
// lector consultaba. Fallaba del lado seguro —nunca salía nada— pero el usuario no podía elegir,
// que es un defecto distinto de estar protegido.
//
// La regla, entonces: **la fuente es el núcleo, y este módulo es el único que pinta**. La
// interfaz pregunta; no recuerda. Sin copia no hay nada que sincronizar, y sin sincronización no
// hay ocasión de olvidarla.

import { t } from "./i18n.js";

const { invoke } = window.__TAURI__.core;
const el = (id) => document.getElementById(id);

/**
 * Pinta las **dos** superficies desde el mismo valor, en la misma llamada.
 *
 * No hay camino por el que una cambie y la otra no, y eso es lo que un test cuenta: un solo
 * escritor por superficie, y el mismo para las dos.
 */
function pintar(local) {
  const insignia = el("mode");
  if (insignia) {
    insignia.dataset.mode = local ? "local" : "hybrid";
    insignia.title = t(local ? "mode.local_hint" : "mode.hybrid_hint");
  }
  const etiqueta = el("mode-label");
  if (etiqueta) etiqueta.textContent = t(local ? "mode.local" : "mode.hybrid");
  const casilla = el("modo-local");
  if (casilla) casilla.checked = local;
  // La consecuencia, en el panel y no en un tooltip: se encontró el control y no se entendió
  // qué decide. Se pinta **aquí**, con las otras dos superficies, para que no pueda discrepar.
  const consecuencia = el("modo-consecuencia");
  if (consecuencia) consecuencia.textContent = t(local ? "mode.local_hint" : "mode.hybrid_hint");
}

/** El modo guardado. Quien lo necesite, que pregunte. */
export async function actual() {
  const dto = await invoke("mode");
  return dto.local;
}

async function refrescar() {
  pintar(await actual());
}

/**
 * Pide el cambio y pinta **lo que el núcleo devuelve**, no lo que se pidió.
 *
 * La diferencia importa: si el núcleo no pudiera aplicarlo, pintar la petición dejaría la
 * pantalla afirmando algo que no ocurrió.
 */
export async function cambiar(local) {
  const dto = await invoke("set_mode", { local });
  pintar(dto.local);
}

/** Repinta al cambiar de idioma. No espera: repintar es idempotente y el valor lo trae el núcleo. */
export function render() {
  refrescar();
}
