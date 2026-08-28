// `003` — conectar proveedores remotos y elegir el modo.
//
// El modo se pinta arriba del panel a propósito: es lo que decide si algo puede salir del
// equipo, y enseñarlo después de conectar sería tarde para decidir.

import { t } from "./i18n.js";
import * as mode from "./mode.js";

const { invoke } = window.__TAURI__.core;

const el = (id) => document.getElementById(id);

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
  const porCodigo = el("desafio-codigo");
  const porCredencial = el("desafio-credencial");
  const sinAlmacen = el("desafio-sin-almacen");
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
    caja.hidden = false;
    if (porCodigo) porCodigo.hidden = desafio.kind !== "delegada";
    if (porCredencial) porCredencial.hidden = desafio.kind !== "credencial";
    if (sinAlmacen) sinAlmacen.hidden = true;
  } catch (err) {
    // FR-004: sin almacén se dice, y se puede seguir en local.
    if (sinAlmacen) {
      caja.hidden = false;
      sinAlmacen.hidden = false;
      sinAlmacen.textContent = `${t("connection.no_store")} (${String(err)})`;
    }
  }
}

/**
 * `003`-FR-008a/b: el modo se guarda y se aplica a la siguiente sesión.
 *
 * Este manejador **pide** el cambio y no pinta nada: quien pinta es `mode.js`, y lo hace con lo
 * que devuelve el núcleo. La casilla que el usuario acaba de pulsar es una petición, no el estado.
 */
async function cambiarModo(evento) {
  await mode.cambiar(evento.target.checked);
  const aviso = el("modo-aviso");
  if (aviso) {
    aviso.hidden = false;
    aviso.textContent = t("mode.changed");
  }
  await refrescar();
}

export function configure() {
  el("conectar")?.addEventListener("click", conectar);
  el("modo-local")?.addEventListener("change", cambiarModo);
  refrescar();
}

/** Repinta lo que se pinta a mano al cambiar de idioma. */
export function render() {
  refrescar();
}
