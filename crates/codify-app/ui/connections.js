// `003` — conectar proveedores remotos y elegir el modo.
//
// El modo se pinta arriba del panel a propósito: es lo que decide si algo puede salir del
// equipo, y enseñarlo después de conectar sería tarde para decidir.

import { t } from "./i18n.js";
import * as mode from "./mode.js";

const { invoke } = window.__TAURI__.core;

const el = (id) => document.getElementById(id);

// El identificador del desafío en curso. **La credencial no se guarda aquí ni en ningún sitio**:
// va del campo al núcleo y no vuelve.
let desafioActual = null;

/** Estado de una conexión, resuelto contra el catálogo (`connection.state.<codigo>`). */
function etiquetaDeEstado(state) {
  return t(`connection.state.${state}`);
}

/** Pinta la lista de cuentas conectadas y quién podría recibir contenido (FR-009). */
export function renderConexiones(conexiones) {
  const lista = el("conexiones-lista");
  const vacio = el("conexiones-vacio");
  if (!lista || !vacio) return;

  lista.replaceChildren(
    ...conexiones.map((c) => {
      const li = document.createElement("li");
      const nombre = document.createElement("span");
      nombre.textContent = `${c.label} — ${etiquetaDeEstado(c.state)}`;
      const boton = document.createElement("button");
      boton.className = "ghost";
      boton.textContent = t("connection.disconnect");
      boton.addEventListener("click", () => desconectar(c.id));
      li.append(nombre, boton);
      return li;
    }),
  );
  vacio.hidden = conexiones.length > 0;

  // FR-009: con remotos permitidos, el usuario ve **antes de empezar** quién podría recibir
  // contenido del repositorio.
  // El modo se pregunta, no se deduce de la casilla: la casilla enseña el modo, no lo define.
  mode.actual().then((local) => {
    const receptores = el("modo-receptores");
    if (!receptores) return;
    const hayRemotos = conexiones.length > 0 && !local;
    receptores.hidden = !hayRemotos;
    if (hayRemotos) {
      receptores.textContent = `${t("mode.will_receive")} ${conexiones
        .map((c) => c.endpointHost || c.label)
        .join(", ")}`;
    }
  });
}

async function refrescar() {
  try {
    renderConexiones(await invoke("list_connections"));
  } catch {
    renderConexiones([]);
  }
}

async function desconectar(id) {
  // SC-006: surte efecto ya, sin reiniciar.
  await invoke("disconnect_provider", { connectionId: id });
  await refrescar();
}

/** Arranca el desafío de conexión y enseña lo que toque según la vía. */
async function conectar() {
  const caja = el("desafio");
  if (!caja) return;

  try {
    // Lo dice el usuario, no el código: sin esto no había forma de conectar a un proveedor
    // concreto, y el endpoint acababa vacío (issue #48).
    const desafio = await invoke("connect_provider", {
      label: el("conn-label")?.value?.trim() || "",
      endpoint: el("conn-endpoint")?.value?.trim() || "",
      tier: el("conn-tier")?.value || "cheap",
      delegada: false,
    });
    abrirDesafio(desafio);
  } catch (err) {
    mostrarFallo("no_store", String(err));
  }
}

/** Enseña el desafío recién abierto, listo para recibir la credencial. */
function abrirDesafio(desafio) {
  desafioActual = desafio.challengeId;
  const caja = el("desafio");
  caja.hidden = false;
  const porCodigo = el("desafio-codigo");
  if (porCodigo) porCodigo.hidden = desafio.kind !== "delegada";

  // La instrucción llega como **clave**, no como frase: el núcleo sabe qué decir, y el idioma lo
  // pone la piel. Antes viajaba redactada, y en español.
  const instruccion = el("desafio-instruccion");
  if (instruccion) {
    instruccion.textContent = desafio.instructions
      ? t(`connection.instruction.${desafio.instructions}`)
      : "";
  }
  const fallo = el("desafio-fallo");
  if (fallo) fallo.hidden = true;
  el("conn-secreto")?.focus();
}

/**
 * Envía la credencial y **cierra el camino**.
 *
 * `complete_connection` existía desde `003` y nadie lo llamaba: el desafío se abría, se enseñaba
 * una frase, y ahí acababa todo. Esta función es la llamada que faltaba.
 */
async function enviarCredencial() {
  const boton = el("desafio-enviar");
  const campo = el("conn-secreto");
  if (!boton || !campo) return;

  // Mientras viaja se dice, y no se admite un segundo envío: pulsar dos veces mandaría el segundo
  // a un desafío ya consumido. Y un envío mudo es indistinguible de que no pase nada.
  boton.disabled = true;
  const textoPrevio = boton.textContent;
  boton.textContent = t("connection.sending");

  try {
    await invoke("complete_connection", {
      challengeId: desafioActual,
      secret: campo.value,
    });
    // La credencial ya cumplió: no se queda en el DOM esperando a que alguien la mire.
    campo.value = "";
    cerrarDesafio();
    await refrescar();
    anunciar(t("connection.connected"));
  } catch (codigo) {
    mostrarFallo(String(codigo));
  } finally {
    boton.disabled = false;
    boton.textContent = textoPrevio;
  }
}

/** Enseña el motivo **y su salida**: un fallo sin salida se parece a que no haya pasado nada. */
function mostrarFallo(codigo, detalle) {
  const fallo = el("desafio-fallo");
  const caja = el("desafio");
  if (!fallo || !caja) return;
  caja.hidden = false;
  fallo.hidden = false;
  const motivo = t(`connection.failure.${codigo}`);
  const salida = t(`connection.failure.${codigo}.next`);
  fallo.textContent = detalle ? `${motivo} ${salida} (${detalle})` : `${motivo} ${salida}`;
  // Los fallos que conservan el desafío dejan el campo listo para corregir sin rehacer nada.
  if (codigo === "rejected" || codigo === "empty") el("conn-secreto")?.focus();
  else cerrarDesafio();
}

function anunciar(texto) {
  const fallo = el("desafio-fallo");
  if (!fallo) return;
  fallo.hidden = false;
  fallo.textContent = texto;
}

/** FR-008: decir «déjalo» y que no quede nada retenido. */
async function cancelarDesafio() {
  await invoke("abandon_connection");
  const campo = el("conn-secreto");
  if (campo) campo.value = "";
  cerrarDesafio();
}

function cerrarDesafio() {
  desafioActual = null;
  const caja = el("desafio");
  if (caja) caja.hidden = true;
}

export function configure() {
  el("conectar")?.addEventListener("click", conectar);
  el("desafio-enviar")?.addEventListener("click", enviarCredencial);
  el("desafio-cancelar")?.addEventListener("click", cancelarDesafio);
  el("modo-local")?.addEventListener("change", cambiarModo);
  refrescar();
}

/** Repinta lo que se pinta a mano al cambiar de idioma. */
export function render() {
  refrescar();
}
