//! Catálogo de cadenas de la interfaz (FR-016b, SC-009).
//!
//! Vive en Rust y no en el frontend por una razón concreta: SC-009 exige **cero cadenas sin
//! traducir**, y eso solo se puede *demostrar* si el catálogo es un dato que un test puede
//! recorrer. Si las cadenas viven sueltas en el DOM, el criterio degrada a una revisión
//! visual — justo lo que la constitución evita.
//!
//! El idioma de la interfaz es **independiente** del de los artefactos generados: alguien con
//! el sistema en inglés puede pedir contexto en español.

use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Locale {
    Es,
    En,
}

impl Locale {
    pub fn code(&self) -> &'static str {
        match self {
            Locale::Es => "es",
            Locale::En => "en",
        }
    }

    /// Interpreta un código de idioma. Cualquier cosa que no sea español cae a inglés.
    pub fn parse(raw: &str) -> Self {
        if raw.trim().to_lowercase().starts_with("es") {
            Locale::Es
        } else {
            Locale::En
        }
    }
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct UiStrings {
    pub locale: Locale,
    pub entries: BTreeMap<&'static str, &'static str>,
}

/// Idioma del sistema, con caída a inglés si no es español.
pub fn system_locale() -> Locale {
    for var in ["LC_ALL", "LC_MESSAGES", "LANG"] {
        if let Ok(value) = std::env::var(var) {
            if !value.trim().is_empty() {
                return Locale::parse(&value);
            }
        }
    }
    Locale::En
}

/// Pares `clave → (español, inglés)`. Una sola fuente para ambos idiomas: es lo que hace
/// **imposible** que una clave exista en uno y falte en el otro.
const CATALOG: &[(&str, &str, &str)] = &[
    // --- Sesión -----------------------------------------------------------
    ("session.repo_placeholder", "Ruta del repositorio a documentar", "Path of the repository to document"),
    ("session.start", "Iniciar", "Start"),
    ("session.cancel", "Cancelar", "Cancel"),
    ("session.cancelling", "Cancelando…", "Cancelling…"),
    ("session.none", "sin sesión", "no session"),
    ("session.state.ingesting", "explorando", "exploring"),
    ("session.state.generating", "generando", "generating"),
    ("session.state.refining", "refinando", "refining"),
    ("session.state.approved", "terminada", "finished"),
    ("session.state.cancelled", "cancelada", "cancelled"),
    ("session.state.failed", "con error", "failed"),
    ("session.interview", "El repositorio está vacío: no hay nada que leer todavía.", "The repository is empty: there is nothing to read yet."),
    // --- Balance de escrituras (FR-017 / FR-023) --------------------------
    ("session.balance.title", "Qué se escribió al repositorio", "What was written to the repository"),
    ("session.balance.none", "No se escribió ningún archivo.", "No file was written."),
    ("session.balance.written", "escrito", "written"),
    ("session.balance.failed", "no se pudo escribir", "could not be written"),
    ("session.balance.skipped", "omitido", "skipped"),
    // --- Corriente de bloques (FR-020) ------------------------------------
    ("stream.hint", "Elige un repositorio y pulsa Iniciar. Verás aquí, en orden, lo que el agente va leyendo y lo que no logra resolver.", "Choose a repository and press Start. You will see here, in order, what the agent reads and what it cannot resolve."),
    ("stream.activity", "actividad", "activity"),
    ("stream.read", "leído", "read"),
    ("stream.unresolved", "no resuelto", "unresolved"),
    ("stream.contradiction", "contradicción entre fuentes", "sources contradict each other"),
    ("stream.written", "escrito", "written"),
    ("stream.egress_blocked", "salida bloqueada", "outbound request blocked"),
    ("stream.error", "error", "error"),
    ("stream.cancelled", "sesión cancelada", "session cancelled"),
    ("stream.omitted.title", "No leído", "Not read"),
    ("stream.omitted.hint", "Quedó fuera del presupuesto de exploración: el resultado no es completo.", "It fell outside the exploration budget: the result is not complete."),
    // --- Proveedor de modelo (FR-019) -------------------------------------
    ("provider.title", "Modelo", "Model"),
    ("provider.checking", "Comprobando el backend…", "Checking the backend…"),
    ("provider.reachable", "Backend disponible", "Backend available"),
    ("provider.unreachable", "No hay backend disponible", "No backend available"),
    ("provider.model_label", "Modelo a usar", "Model to use"),
    ("provider.no_models", "El backend responde pero no tiene modelos instalados.", "The backend responds but has no models installed."),
    ("provider.retry", "Volver a comprobar", "Check again"),
    ("provider.next_step", "Qué hacer", "What to do"),
    // --- Idioma (FR-016 / FR-016b) ----------------------------------------
    ("locale.ui_label", "Idioma de la interfaz", "Interface language"),
    ("locale.artifacts_label", "Idioma del contexto generado", "Generated context language"),
    ("locale.independent_hint", "Son independientes: puedes usar la aplicación en un idioma y generar el contexto en otro.", "They are independent: you can use the app in one language and generate the context in another."),
    // --- Modo (FR-005) ----------------------------------------------------
    ("mode.local", "modo local", "local mode"),
    ("mode.hybrid", "modo híbrido", "hybrid mode"),
    ("mode.local_hint", "Nada del repositorio sale hacia servicios de nube.", "Nothing from the repository leaves towards cloud services."),
    ("mode.hybrid_hint", "Se pueden consultar servicios remotos que hayas conectado.", "Remote services you have connected may be queried."),
    // --- Artefactos (FR-011 a FR-013, FR-021) -----------------------------
    ("artifact.open", "Ver archivo completo", "View full file"),
    ("artifact.close", "Cerrar", "Close"),
    ("artifact.grounded", "verificado", "verified"),
    ("artifact.tentative", "sin verificar", "unverified"),
    ("artifact.contradiction", "fuentes en conflicto", "conflicting sources"),
    ("artifact.sources", "Fuente", "Source"),
    ("artifact.reason", "Motivo", "Reason"),
    ("artifact.not_written", "todavía no está en el repositorio", "not in the repository yet"),
    ("artifact.pending_tentative", "Hay puntos sin verificar. Puedes resolverlos o dejarlos declarados como pendientes.", "There are unverified points. You can resolve them or leave them declared as pending."),
    // --- Errores accionables (FR-028) -------------------------------------
    ("error.no_repo", "Elige un repositorio antes de iniciar.", "Choose a repository before starting."),
    ("error.session_failed", "La sesión no pudo completarse", "The session could not be completed"),
    ("error.unknown", "Ocurrió algo inesperado", "Something unexpected happened"),
    // --- Cierre (FR-024) --------------------------------------------------
    ("close.title", "Hay una sesión en curso", "A session is in progress"),
    ("close.warning", "Si cierras ahora, la sesión termina y no se reanuda. Lo ya escrito permanece en el repositorio.", "If you close now, the session ends and will not resume. What has already been written stays in the repository."),
    ("close.confirm", "Cerrar de todos modos", "Close anyway"),
    ("close.cancel", "Seguir trabajando", "Keep working"),
    // --- Accesibilidad (FR-025 a FR-027) ----------------------------------
    ("a11y.stream_region", "Actividad del agente", "Agent activity"),
    ("a11y.artifact_region", "Contenido del archivo", "File content"),
    ("a11y.provider_region", "Configuración del modelo", "Model configuration"),
    ("a11y.toolbar_region", "Controles de la sesión", "Session controls"),
    ("a11y.shortcuts", "Atajos: Enter inicia, Escape cancela, flechas recorren la actividad.", "Shortcuts: Enter starts, Escape cancels, arrows move through the activity."),
];

pub fn strings_for(locale: Locale) -> UiStrings {
    let entries = CATALOG
        .iter()
        .map(|(key, es, en)| {
            let value = match locale {
                Locale::Es => *es,
                Locale::En => *en,
            };
            (*key, value)
        })
        .collect();
    UiStrings { locale, entries }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// **SC-009**: cero cadenas sin traducir. La paridad es estructural —el catálogo declara
    /// ambos idiomas en la misma fila— pero el test lo verifica igual: si alguien cambia la
    /// forma del catálogo, el build lo dice.
    #[test]
    fn both_locales_expose_exactly_the_same_keys() {
        let es = strings_for(Locale::Es);
        let en = strings_for(Locale::En);

        let es_keys: Vec<_> = es.entries.keys().collect();
        let en_keys: Vec<_> = en.entries.keys().collect();
        assert_eq!(
            es_keys, en_keys,
            "las claves deben coincidir en ambos idiomas"
        );
    }

    #[test]
    fn no_entry_is_empty_in_any_locale() {
        for locale in [Locale::Es, Locale::En] {
            for (key, value) in strings_for(locale).entries {
                assert!(
                    !value.trim().is_empty(),
                    "la clave '{key}' está vacía en {}",
                    locale.code()
                );
            }
        }
    }

    #[test]
    fn there_are_no_duplicate_keys() {
        let mut seen = std::collections::HashSet::new();
        for (key, _, _) in CATALOG {
            assert!(seen.insert(*key), "clave duplicada: {key}");
        }
    }

    /// Cada superficie de la interfaz tiene cadenas: si se olvida un espacio de nombres
    /// completo, algo quedaría sin traducir sin que nadie lo note.
    #[test]
    fn every_expected_namespace_is_present() {
        let entries = strings_for(Locale::Es).entries;
        for prefix in [
            "session.",
            "stream.",
            "provider.",
            "locale.",
            "mode.",
            "artifact.",
            "error.",
            "close.",
            "a11y.",
        ] {
            assert!(
                entries.keys().any(|k| k.starts_with(prefix)),
                "falta el espacio de nombres '{prefix}'"
            );
        }
    }

    #[test]
    fn spanish_and_english_actually_differ() {
        let es = strings_for(Locale::Es);
        let en = strings_for(Locale::En);
        // No es una traducción de mentira: la mayoría de los textos cambian.
        let identical = es
            .entries
            .iter()
            .filter(|(k, v)| en.entries.get(*k) == Some(v))
            .count();
        assert!(
            identical * 4 < es.entries.len(),
            "demasiadas cadenas idénticas ({identical}/{}): ¿se tradujo de verdad?",
            es.entries.len()
        );
    }

    /// Extrae las claves de traducción de un archivo de la interfaz.
    ///
    /// Se buscan **solo** las formas que consumen el catálogo: `t("clave")` en JavaScript y
    /// los atributos `data-i18n*` en HTML. Reconocer cualquier cadena con punto sería
    /// demasiado laxo — confundiría los nombres de evento de la ventana con traducciones.
    fn keys_used_in(content: &str, is_html: bool) -> Vec<String> {
        let mut keys = Vec::new();

        let mut collect = |marker: &str, split_on_colon: bool| {
            let mut rest = content;
            while let Some(at) = rest.find(marker) {
                // Para `t("`, el carácter previo no puede ser parte de un identificador,
                // o se colarían cosas como `format("`.
                let prefix_ok = !marker.starts_with("t(") || {
                    let before = rest[..at].chars().last();
                    !matches!(before, Some(c) if c.is_alphanumeric() || c == '_')
                };
                rest = &rest[at + marker.len()..];
                if !prefix_ok {
                    continue;
                }
                if let Some(end) = rest.find('"') {
                    for part in rest[..end].split(',') {
                        let value = if split_on_colon {
                            part.split(':').nth(1).unwrap_or(part)
                        } else {
                            part
                        };
                        let value = value.trim();
                        if value.contains('.') {
                            keys.push(value.to_string());
                        }
                    }
                }
            }
        };

        if is_html {
            collect("data-i18n=\"", false);
            collect("data-i18n-aria=\"", false);
            collect("data-i18n-attr=\"", true);
        } else {
            collect("t(\"", false);
        }
        keys
    }

    /// Cierra el hueco que deja "el frontend no tiene tests automatizados": recorre los
    /// archivos de la interfaz y comprueba que **toda clave que usan existe en el catálogo**.
    /// Sin esto, una clave mal escrita se vería como texto crudo en pantalla y solo lo
    /// notaría alguien mirando. Con esto, lo dice el build.
    #[test]
    fn every_key_used_by_the_ui_exists_in_the_catalog() {
        let ui_dir = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("ui");
        let known: std::collections::HashSet<&str> = CATALOG.iter().map(|(k, _, _)| *k).collect();

        let mut checked = 0usize;
        let mut missing = Vec::new();

        for entry in std::fs::read_dir(&ui_dir).expect("falta el directorio ui/") {
            let path = entry.unwrap().path();
            let ext = path.extension().and_then(|e| e.to_str()).unwrap_or("");
            if !matches!(ext, "js" | "html") {
                continue;
            }
            let content = std::fs::read_to_string(&path).unwrap();
            let name = path.file_name().unwrap().to_string_lossy().to_string();

            for key in keys_used_in(&content, ext == "html") {
                checked += 1;
                if !known.contains(key.as_str()) {
                    missing.push(format!("{name}: {key}"));
                }
            }
        }

        assert!(
            checked > 25,
            "se esperaban muchas claves en la UI, se hallaron {checked}"
        );
        assert!(
            missing.is_empty(),
            "claves usadas por la UI que no están en el catálogo:\n{}",
            missing.join("\n")
        );
    }

    #[test]
    fn locale_parsing_falls_back_to_english() {
        assert_eq!(Locale::parse("es_CO.UTF-8"), Locale::Es);
        assert_eq!(Locale::parse("es"), Locale::Es);
        assert_eq!(Locale::parse("en_US.UTF-8"), Locale::En);
        assert_eq!(
            Locale::parse("fr_FR"),
            Locale::En,
            "lo desconocido cae a inglés"
        );
        assert_eq!(Locale::parse(""), Locale::En);
    }
}
