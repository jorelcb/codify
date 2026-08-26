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
    ("session.state.interview", "sin material", "no material"),
    ("session.interview", "El repositorio está vacío: no hay nada que leer todavía.", "The repository is empty: there is nothing to read yet."),
    ("session.interview_next", "No hay nada que leer, así que nada puede fundamentarse en una fuente. Descríbelo abajo en tus palabras: se redactará a partir de lo que digas, marcado como sin verificar hasta que exista código que lo respalde.", "There is nothing to read, so nothing can be grounded in a source. Describe it below in your own words: it will be drafted from what you say, marked as unverified until there is code to back it up."),
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
    ("stream.interview", "sin material que leer", "no material to read"),
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
    // `003` — conectar cuentas remotas y elegir el modo.
    ("connection.title", "Proveedores conectados", "Connected providers"),
    ("connection.add", "Conectar un proveedor", "Connect a provider"),
    ("connection.none", "Ninguno. La aplicación funciona en local.", "None. The app runs locally."),
    ("connection.disconnect", "Desconectar", "Disconnect"),
    ("connection.code_hint", "Introduce este código en la página que se abrirá:", "Enter this code on the page that will open:"),
    ("connection.secret_hint", "Se guarda en el almacén del sistema y no vuelve a mostrarse.", "It is stored in the system keychain and never shown again."),
    ("connection.no_store", "Este sistema no tiene almacén de credenciales. Puedes seguir en modo local.", "This system has no credential store. You can continue in local mode."),
    ("connection.state.connected", "conectada", "connected"),
    ("connection.state.expired", "caducada: hay que reconectar", "expired: reconnect needed"),
    ("connection.state.revoked", "revocada", "revoked"),
    ("mode.title", "Modo", "Mode"),
    ("mode.local_option", "Solo local — nada sale de este equipo", "Local only — nothing leaves this machine"),
    ("mode.changed", "El modo nuevo se aplica a la siguiente sesión; la actual termina como empezó.", "The new mode applies to the next session; the current one finishes as it started."),
    ("mode.will_receive", "Con este modo, estos proveedores pueden recibir contenido del repositorio:", "In this mode, these providers may receive repository content:"),
    // `001`-FR-018: degradar sin decirlo entrega calidad reducida haciéndola pasar por la buena.
    (
        "provider.tier_degraded",
        "Generado con calidad reducida: no había un modelo del nivel pedido y se usó el disponible.",
        "Generated with reduced quality: no model of the requested tier was available, so the available one was used.",
    ),
    // Un texto por cada `ProviderIssue::code()`. El núcleo nombra el motivo; el idioma
    // se decide aquí. Hay un test que comprueba que no falte ninguno.
    ("provider.issue.no_models", "Descarga un modelo, por ejemplo `ollama pull qwen2.5-coder`.", "Download a model, for example `ollama pull qwen2.5-coder`."),
    ("provider.issue.not_listening", "Arranca un backend local, por ejemplo con `ollama serve`, o apunta a otro endpoint.", "Start a local backend, for example with `ollama serve`, or point to another endpoint."),
    ("provider.issue.endpoint_not_local", "En modo local el endpoint tiene que estar en esta máquina: usa una dirección de loopback.", "In local mode the endpoint must be on this machine: use a loopback address."),
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
    ("artifact.title", "Archivo generado", "Generated file"),
    ("artifact.select_label", "Archivo", "File"),
    ("artifact.none", "Todavía no se ha generado ningún archivo.", "No file has been generated yet."),
    ("artifact.empty", "Este archivo todavía no tiene contenido.", "This file has no content yet."),
    ("artifact.grounded", "verificado", "verified"),
    ("artifact.tentative", "sin verificar", "unverified"),
    ("artifact.contradiction", "fuentes en conflicto", "conflicting sources"),
    ("artifact.sources", "Fuente", "Source"),
    ("artifact.reason", "Motivo", "Reason"),
    // Estado de escritura: un archivo en pantalla no es un archivo en el repositorio (FR-017).
    ("artifact.not_written", "todavía no está en el repositorio", "not in the repository yet"),
    ("artifact.in_repository", "está en el repositorio", "in the repository"),
    ("artifact.write_failed", "no se pudo escribir en el repositorio", "could not be written to the repository"),
    ("artifact.write_skipped", "no se escribió", "not written"),
    // Diferir (FR-014): decidir dejarlo pendiente, a la vista, no despacharlo.
    ("artifact.defer", "Dejar pendiente", "Leave pending"),
    ("artifact.deferred", "pendiente, a sabiendas", "pending, knowingly"),
    ("artifact.defer_hint", "Dejarlo pendiente no lo borra ni lo da por bueno: queda declarado como sin verificar en el archivo.", "Leaving it pending neither deletes it nor accepts it: it stays declared as unverified in the file."),
    ("artifact.pending_tentative", "Hay puntos sin verificar. Puedes resolverlos o dejarlos declarados como pendientes.", "There are unverified points. You can resolve them or leave them declared as pending."),
    // --- Refinamiento conversacional y decisión sobre propuestas (US2 de 002) -----
    ("refine.placeholder", "Corrige o pide un cambio en tus palabras…", "Correct something or ask for a change in your own words…"),
    ("refine.send", "Enviar", "Send"),
    ("refine.thinking", "Preparando propuestas…", "Preparing proposals…"),
    ("refine.no_changes", "No hubo nada que cambiar con eso.", "Nothing to change from that."),
    ("proposal.label", "propuesta de cambio", "proposed change"),
    ("proposal.title", "Decidir sobre este cambio", "Decide on this change"),
    ("proposal.rationale", "Por qué", "Why"),
    ("proposal.approve", "Aprobar", "Approve"),
    ("proposal.edit", "Editar", "Edit"),
    ("proposal.reject", "Rechazar", "Reject"),
    ("proposal.edit_hint", "Escribe cómo debería quedar. Se aplicará tu texto, no el del agente.", "Write how it should read. Your text will be applied, not the agent's."),
    ("proposal.pending", "sin decidir", "undecided"),
    ("proposal.prev", "Anterior", "Previous"),
    ("proposal.next", "Siguiente", "Next"),
    ("proposal.decided", "decidida", "decided"),
    ("proposal.approved", "aprobada", "approved"),
    ("proposal.edited", "aplicada con tu texto", "applied with your text"),
    ("proposal.rejected", "rechazada: el archivo no cambió", "rejected: the file did not change"),
    ("proposal.auto_applied", "aplicada sin preguntar (bajo riesgo)", "applied without asking (low risk)"),
    ("proposal.revert", "Deshacer", "Undo"),
    ("proposal.reverted", "deshecha: el archivo volvió a como estaba", "undone: the file went back to how it was"),
    ("proposal.applied_title", "Aplicados sin preguntar", "Applied without asking"),
    // --- Errores accionables (FR-028) -------------------------------------
    ("error.no_repo", "Elige un repositorio antes de iniciar.", "Choose a repository before starting."),
    ("error.session_failed", "La sesión no pudo completarse", "The session could not be completed"),
    ("error.unknown", "Ocurrió algo inesperado", "Something unexpected happened"),
    // `002`-FR-028: cada motivo lleva **qué pasó** y **qué hacer**. Sin lo segundo, saber por
    // qué falló deja al usuario igual de atascado.
    ("session.failure.provider_timeout", "El modelo tardó más de lo permitido.", "The model took longer than allowed."),
    ("session.failure.provider_timeout.next", "Prueba con un modelo más pequeño, o vuelve a intentarlo: los modelos grandes tardan minutos por artefacto.", "Try a smaller model, or run it again: large models take minutes per artifact."),
    ("session.failure.provider_unavailable", "No se pudo contactar con el backend de modelos.", "The model backend could not be reached."),
    ("session.failure.provider_unavailable.next", "Comprueba que esté levantado y escuchando en el endpoint configurado.", "Check that it is running and listening on the configured endpoint."),
    ("session.failure.provider_unparseable", "El modelo respondió algo que no se pudo interpretar.", "The model returned something that could not be interpreted."),
    ("session.failure.provider_unparseable.next", "Suele pasar con modelos pequeños. Prueba con uno mayor.", "This usually happens with small models. Try a larger one."),
    ("session.failure.repo_unreadable", "No se pudo leer el repositorio o escribir en él.", "The repository could not be read or written to."),
    ("session.failure.repo_unreadable.next", "Comprueba la ruta y los permisos del directorio.", "Check the path and the directory permissions."),
    ("session.failure.egress_blocked", "La política de cero-egress cortó la operación.", "The zero-egress policy stopped the operation."),
    ("session.failure.egress_blocked.next", "En modo local es lo esperado: nada sale del equipo. Revisa qué se intentó en el registro.", "In local mode this is expected: nothing leaves your machine. Check the log to see what was attempted."),
    ("session.failure.unauthorized", "Falta autorización para algo que se intentó.", "Authorization is missing for something that was attempted."),
    ("session.failure.unauthorized.next", "Revisa las credenciales del proveedor configurado.", "Check the credentials of the configured provider."),
    ("session.failure.internal", "La sesión falló por un problema interno.", "The session failed due to an internal problem."),
    ("session.failure.internal.next", "Vuelve a intentarlo. Si se repite, el detalle técnico está en el registro de actividad.", "Try again. If it repeats, the technical detail is in the activity log."),
    // --- Cierre (FR-024) --------------------------------------------------
    ("close.title", "Hay una sesión en curso", "A session is in progress"),
    ("close.warning", "Si cierras ahora, la sesión termina y no se reanuda. Lo ya escrito permanece en el repositorio.", "If you close now, the session ends and will not resume. What has already been written stays in the repository."),
    ("close.confirm", "Cerrar de todos modos", "Close anyway"),
    ("close.cancel", "Seguir trabajando", "Keep working"),
    // Cierre con puntos sin verificar (FR-014). No se bloquea al usuario: se le hace decidir.
    ("close.tentative_title", "Hay puntos sin verificar", "There are unverified points"),
    ("close.tentative_warning", "Quedan puntos sin verificar y sin atender. Si cierras ahora se quedan declarados como pendientes en el archivo: no se pierden, pero nadie los ha mirado.", "Some points remain unverified and unattended. If you close now they stay declared as pending in the file: nothing is lost, but nobody has looked at them."),
    ("close.review", "Revisarlos", "Review them"),
    // --- Accesibilidad (FR-025 a FR-027) ----------------------------------
    ("a11y.stream_region", "Actividad del agente", "Agent activity"),
    ("a11y.artifact_region", "Contenido del archivo", "File content"),
    ("a11y.decide_region", "Cambios que esperan tu decisión", "Changes awaiting your decision"),
    ("a11y.applied_region", "Cambios aplicados sin preguntar", "Changes applied without asking"),
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
