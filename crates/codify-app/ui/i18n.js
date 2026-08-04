// Consumo del catálogo de cadenas (FR-016b).
//
// Ninguna cadena visible se escribe aquí: todas vienen del catálogo en Rust, que es donde un
// test puede verificar que no falta ninguna traducción (SC-009). Este módulo solo las aplica
// al DOM.

const { invoke } = window.__TAURI__.core;

let entries = {};
let current = "es";

/** Texto de una clave. Si faltara, se ve la clave — nunca una cadena vacía silenciosa. */
export function t(key) {
  return entries[key] ?? key;
}

export function locale() {
  return current;
}

/** Carga el catálogo y repinta toda la interfaz. Cambiar idioma no exige reiniciar. */
export async function setLocale(code) {
  const catalog = await invoke("ui_strings", { locale: code });
  entries = catalog.entries;
  current = catalog.locale;
  document.documentElement.lang = current;
  apply();
  return current;
}

/** Idioma del sistema, con caída a inglés. */
export async function detectLocale() {
  return invoke("system_locale");
}

/** Aplica el catálogo a todo el DOM marcado con atributos `data-i18n*`. */
export function apply(root = document) {
  root.querySelectorAll("[data-i18n]").forEach((el) => {
    el.textContent = t(el.dataset.i18n);
  });

  // `data-i18n-attr="placeholder:clave"` para atributos.
  root.querySelectorAll("[data-i18n-attr]").forEach((el) => {
    for (const pair of el.dataset.i18nAttr.split(",")) {
      const [attr, key] = pair.split(":");
      el.setAttribute(attr.trim(), t(key.trim()));
    }
  });

  // Nombres de región para lectores de pantalla (FR-027).
  root.querySelectorAll("[data-i18n-aria]").forEach((el) => {
    el.setAttribute("aria-label", t(el.dataset.i18nAria));
  });
}
