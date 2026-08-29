//! Contrato de la interfaz: lo que el quickstart pedía comprobar **a ojo**, comprobado por el
//! build.
//!
//! Los escenarios S5–S8 estaban escritos como inspección manual («recorrer la aplicación en
//! ambos idiomas buscando claves crudas», «verificar el orden de tabulación»). Marcarlos como
//! hechos por haberlos mirado una vez es exactamente el tipo de afirmación que la constitución
//! rechaza: no se puede verificar contra una fuente, y se pudre en el siguiente cambio de UI.
//!
//! Estos tests no sustituyen al recorrido humano —nadie puede automatizar «se entiende»— pero
//! sí clavan las propiedades que sí son mecánicas, y que son justo las que se rompen solas.

use codify_app::strings::{strings_for, Locale};
use std::path::{Path, PathBuf};

fn crate_dir() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn ui_dir() -> PathBuf {
    crate_dir().join("ui")
}

fn read(path: impl AsRef<Path>) -> String {
    let path = path.as_ref();
    std::fs::read_to_string(path).unwrap_or_else(|e| panic!("no se pudo leer {path:?}: {e}"))
}

/// Todos los archivos de la interfaz, como `(nombre, contenido)`.
fn ui_files() -> Vec<(String, String)> {
    let mut out = Vec::new();
    for entry in std::fs::read_dir(ui_dir()).expect("falta el directorio ui/") {
        let path = entry.unwrap().path();
        let ext = path.extension().and_then(|e| e.to_str()).unwrap_or("");
        if matches!(ext, "js" | "html" | "css") {
            let name = path.file_name().unwrap().to_string_lossy().to_string();
            out.push((name, read(&path)));
        }
    }
    assert!(out.len() >= 5, "se esperaban los archivos de la interfaz");
    out
}

// ---------------------------------------------------------------------------
// S7 · Ventana mínima
// ---------------------------------------------------------------------------

/// El punto de quiebre responsivo tiene que ser **alcanzable**.
///
/// Este test nace de un defecto real: el CSS declaraba reglas para «ventana pequeña» bajo
/// `max-width: 720px` mientras la ventana no podía bajar de 820px. Nunca se aplicaban. Código
/// muerto que *aparentaba* cubrir SC-007 — la peor clase, porque se lee como una garantía.
#[test]
fn el_punto_de_quiebre_responsivo_es_alcanzable() {
    let conf = read(crate_dir().join("tauri.conf.json"));
    let min_width = conf
        .split("\"minWidth\"")
        .nth(1)
        .and_then(|rest| rest.split_once(':'))
        .map(|(_, value)| value.split(',').next().unwrap_or(value))
        .and_then(|value| value.trim().parse::<u32>().ok())
        .expect("tauri.conf.json debe declarar minWidth");

    // Hay que leer `max-width` **de la condición** de la media query, es decir antes de su
    // `{`. Buscarlo a secas encuentra antes el ancho máximo de la corriente
    // (`.stream { max-width: 920px }`), que no es un punto de quiebre.
    let css = read(ui_dir().join("styles.css"));
    let breakpoints: Vec<u32> = css
        .split("@media")
        .skip(1)
        .filter_map(|rest| rest.split('{').next())
        .filter_map(|cond| cond.split_once("max-width:"))
        .filter_map(|(_, value)| value.trim().split("px").next())
        .filter_map(|n| n.trim().parse::<u32>().ok())
        .collect();

    assert!(
        !breakpoints.is_empty(),
        "styles.css debe declarar al menos un `@media (max-width: Npx)`"
    );

    let muertos: Vec<u32> = breakpoints
        .iter()
        .copied()
        .filter(|bp| *bp < min_width)
        .collect();
    assert!(
        muertos.is_empty(),
        "la ventana no baja de {min_width}px, así que las reglas de `max-width` en {muertos:?} \
         nunca se aplican: o se baja el mínimo, o son código muerto que aparenta cubrir SC-007"
    );
}

// ---------------------------------------------------------------------------
// S5 · Idioma — cero cadenas fuera del catálogo
// ---------------------------------------------------------------------------

/// Texto que se muestra tal cual y **no** viene del catálogo: se quedaría en un solo idioma.
///
/// Se permiten solo estos, y cada uno por un motivo:
/// - `codify` es el nombre del producto, no se traduce.
/// - `Español` / `English` son endónimos: cada idioma se nombra **en sí mismo** en el selector,
///   que es justo lo que permite encontrarlo cuando la interfaz está en el idioma que no
///   entiendes.
const TEXTO_LITERAL_PERMITIDO: &[&str] = &["codify", "Español", "English"];

/// Borra todo lo que va entre `open` y `close`, inclusive.
fn strip_between(source: &str, open: &str, close: &str) -> String {
    let mut out = String::with_capacity(source.len());
    let mut cursor = source;
    while let Some(at) = cursor.find(open) {
        out.push_str(&cursor[..at]);
        match cursor[at..].find(close) {
            Some(end) => cursor = &cursor[at + end + close.len()..],
            None => return out, // sin cierre: se descarta el resto
        }
    }
    out.push_str(cursor);
    out
}

/// Extrae los pares `(etiqueta de apertura, texto que contiene)` del HTML.
fn text_runs(html: &str) -> Vec<(String, String)> {
    // Fuera comentarios y bloques que no son contenido visible.
    let clean = strip_between(html, "<!--", "-->");
    let clean = strip_between(&clean, "<script", "</script>");
    let clean = strip_between(&clean, "<style", "</style>");

    let mut runs = Vec::new();
    let mut cursor = clean.as_str();
    while let Some(open_at) = cursor.find('<') {
        cursor = &cursor[open_at..];
        let Some(close_at) = cursor.find('>') else {
            break;
        };
        let tag = cursor[..=close_at].to_string();
        cursor = &cursor[close_at + 1..];
        let text: String = cursor.chars().take_while(|c| *c != '<').collect();
        if !text.trim().is_empty() {
            runs.push((tag, text.trim().to_string()));
        }
    }
    runs
}

#[test]
fn ningun_texto_visible_escapa_al_catalogo() {
    let html = read(ui_dir().join("index.html"));
    let mut sueltos = Vec::new();

    for (tag, text) in text_runs(&html) {
        let traducido = tag.contains("data-i18n");
        let decorativo = tag.contains("aria-hidden=\"true\"");
        let permitido = TEXTO_LITERAL_PERMITIDO.contains(&text.as_str());
        if !traducido && !decorativo && !permitido {
            sueltos.push(format!("{text:?} dentro de {}", tag.trim()));
        }
    }

    assert!(
        sueltos.is_empty(),
        "texto escrito directamente en el HTML: se quedaría en un solo idioma (SC-009).\n{}",
        sueltos.join("\n")
    );
}

/// Un elemento no puede tener **dos dueños** de su texto.
///
/// Este test nace de otro defecto real: `#provider-status` llevaba `data-i18n` *y* lo escribía
/// `provider.js` a mano. Al cambiar de idioma, `i18n.apply()` lo devolvía a «comprobando el
/// backend…» mientras el glifo seguía en ✓ — el panel afirmaba dos cosas contradictorias, y
/// mentía sobre si había backend.
///
/// La regla: si el JavaScript le escribe el `textContent`, el HTML no lleva `data-i18n`. El
/// repintado al cambiar de idioma es entonces responsabilidad explícita de quien lo pinta.
#[test]
fn ningun_elemento_tiene_dos_duenos_de_su_texto() {
    let html = read(ui_dir().join("index.html"));

    // Ids cuyo texto pinta `apply()`.
    let traducidos: Vec<String> = text_runs_tags(&html)
        .into_iter()
        .filter(|tag| tag.contains("data-i18n=\""))
        .filter_map(|tag| between(&tag, "id=\"", "\""))
        .collect();

    // El análisis va **archivo a archivo**: la atadura entre una variable y su id es local al
    // módulo. Analizar todo el JavaScript junto daba falsos positivos —`nextEl` es
    // `provider-next` en un módulo y `decide-next` en otro—, y un test que grita sin motivo
    // acaba ignorándose, que es peor que no tenerlo.
    let mut conflictos = Vec::new();
    for (name, content) in ui_files() {
        if !name.ends_with(".js") {
            continue;
        }
        let mut rest = content.as_str();
        while let Some(at) = rest.find("document.getElementById(\"") {
            let nombre: String = rest[..at]
                .trim_end()
                .trim_end_matches(['=', ':'])
                .trim_end()
                .chars()
                .rev()
                .take_while(|c| c.is_alphanumeric() || *c == '_' || *c == '$')
                .collect::<Vec<_>>()
                .into_iter()
                .rev()
                .collect();
            rest = &rest[at + "document.getElementById(\"".len()..];
            let Some(id) = rest.split('"').next().map(str::to_string) else {
                continue;
            };
            if nombre.is_empty() || !traducidos.contains(&id) {
                continue;
            }
            if content.contains(&format!("{nombre}.textContent =")) {
                conflictos.push(format!(
                    "{name} #{id}: lleva `data-i18n` en el HTML y `{nombre}.textContent =` en el \
                     JavaScript"
                ));
            }
        }
    }

    assert!(
        conflictos.is_empty(),
        "elementos con dos dueños de su texto: al cambiar de idioma se pisan entre sí.\n{}",
        conflictos.join("\n")
    );
}

/// Módulos cuyo texto **no** se repinta al cambiar de idioma, y por qué está bien.
///
/// `stream.js` es append-only: sus bloques espejan el log de auditoría del núcleo. Reescribir
/// un bloque ya emitido sería falsificar lo que pasó, así que se queda en el idioma en que
/// ocurrió — a propósito.
const NO_SE_REPINTAN: &[&str] = &["stream.js"];

/// Quien pinta texto a mano tiene que **repintarlo** al cambiar de idioma.
///
/// Este test nace del tercer defecto de esta familia. `i18n.apply()` solo alcanza al DOM
/// marcado con `data-i18n`; todo lo que un módulo escriba con `textContent` se queda en el
/// idioma en que se pintó. Pasó con «sin sesión», pasó con el panel del proveedor, y volvió a
/// pasar con las etiquetas de fundamento de la vista de artefacto — que son justo lo que esa
/// vista existe para comunicar.
///
/// La regla: si un módulo usa el catálogo, expone `render()` y el manejador de cambio de
/// idioma lo llama.
#[test]
fn quien_pinta_a_mano_repinta_al_cambiar_de_idioma() {
    let main_js = read(ui_dir().join("main.js"));

    // Cuerpo del manejador de cambio de idioma de la interfaz.
    let handler = main_js
        .split_once("el.uiLocale.addEventListener")
        .map(|(_, rest)| rest.split("\n});").next().unwrap_or(rest))
        .expect("main.js debe reaccionar al cambio de idioma de la interfaz");

    let mut sin_repintar = Vec::new();
    for (name, content) in ui_files() {
        if !name.ends_with(".js") || matches!(name.as_str(), "main.js" | "i18n.js") {
            continue;
        }
        if NO_SE_REPINTAN.contains(&name.as_str()) {
            continue;
        }
        // ¿Consume el catálogo? Entonces pinta texto que hay que repintar.
        if !content.contains("t(\"") {
            continue;
        }
        let module = name.trim_end_matches(".js");
        if !content.contains("export function render") {
            sin_repintar.push(format!("{name}: no expone `render()`"));
        } else if !handler.contains(&format!("{module}.render(")) {
            sin_repintar.push(format!(
                "{name}: expone `render()` pero el cambio de idioma no lo llama"
            ));
        }
    }

    assert!(
        sin_repintar.is_empty(),
        "texto que se quedaría en el idioma anterior al cambiar de idioma (SC-009).\n{}",
        sin_repintar.join("\n")
    );
}

/// Todas las etiquetas de apertura del HTML, ya sin comentarios ni scripts.
fn text_runs_tags(html: &str) -> Vec<String> {
    let clean = strip_between(html, "<!--", "-->");
    let clean = strip_between(&clean, "<script", "</script>");
    let mut tags = Vec::new();
    let mut cursor = clean.as_str();
    while let Some(open_at) = cursor.find('<') {
        cursor = &cursor[open_at..];
        let Some(close_at) = cursor.find('>') else {
            break;
        };
        tags.push(cursor[..=close_at].to_string());
        cursor = &cursor[close_at + 1..];
    }
    tags
}

fn between(source: &str, open: &str, close: &str) -> Option<String> {
    let (_, rest) = source.split_once(open)?;
    rest.split(close).next().map(str::to_string)
}

/// Claves del catálogo que **nadie usa**: o falta cablearlas, o sobran.
///
/// Aquí vivían las nueve claves `artifact.*` reservadas de antemano para la vista de artefacto
/// completo (US3, issue #6). La lista quedó **vacía** cuando esa vista aterrizó, que era
/// exactamente lo que debía pasar: si hubiera sobrevivido, habría significado que la vista se
/// escribió con texto propio en lugar de consumir el catálogo.
///
/// Mantenerla vacía es lo correcto. Volver a llenarla solo se justifica reservando claves para
/// una superficie que todavía no existe — y con el issue que la va a consumir escrito al lado.
const RESERVADAS: &[&str] = &[];

#[test]
fn toda_clave_del_catalogo_esta_cableada_o_declarada_como_reservada() {
    let files = ui_files();
    let rust = read(crate_dir().join("src/commands.rs"));
    let mut huerfanas = Vec::new();

    for key in strings_for(Locale::Es).entries.keys() {
        if RESERVADAS.contains(key) {
            continue;
        }
        // Las claves compuestas se arman en tiempo de ejecución (`session.state.${state}`), así
        // que no aparecen literalmente. Pero exigir solo que el **prefijo** esté en algún sitio
        // era demasiado laxo: `provider.tier_degraded` pasaba por el mero hecho de que la
        // palabra «provider» saliera en `provider.issue.${…}`, y una clave sin cablear se daba
        // por consumida. Ahora se exige la plantilla concreta: `<prefijo>.${`.
        // Se prueban **todos** los cortes, no solo el último: una clave compuesta puede llevar
        // sufijo tras la interpolación (`session.failure.${motivo}.next`), y probar un único
        // prefijo la daba por huérfana. Sigue exigiendo la plantilla concreta, que es lo que
        // impide que valga cualquier mención suelta del prefijo.
        let compuesta = key.match_indices('.').any(|(i, _)| {
            let plantilla = format!("{}.${{", &key[..i]);
            files.iter().any(|(_, c)| c.contains(&plantilla))
        });
        let usada = files.iter().any(|(_, c)| c.contains(key)) || rust.contains(key) || compuesta;
        if !usada {
            huerfanas.push(*key);
        }
    }

    assert!(
        huerfanas.is_empty(),
        "claves del catálogo que nadie consume (¿falta cablearlas?): {huerfanas:?}"
    );
}

/// Cada motivo que el núcleo sabe reportar tiene texto en **los dos** idiomas.
///
/// El núcleo devuelve un código (`ProviderIssue::code()`) y la piel elige la frase. Ese
/// desacople es lo que permite que SC-009 sea demostrable —el núcleo ya no redacta en un
/// idioma fijo—, pero abre un hueco nuevo: un motivo sin entrada en el catálogo se vería en
/// pantalla como `provider.issue.loquesea`. Esto lo cierra.
#[test]
fn todo_motivo_del_proveedor_tiene_texto_en_ambos_idiomas() {
    use codify_core::application::ports::ProviderIssue;

    let motivos = [
        ProviderIssue::NoModels,
        ProviderIssue::NotListening,
        ProviderIssue::EndpointNotLocal,
    ];

    for locale in [Locale::Es, Locale::En] {
        let entries = strings_for(locale).entries;
        for motivo in motivos {
            let key = format!("provider.issue.{}", motivo.code());
            let texto = entries.get(key.as_str()).unwrap_or_else(|| {
                panic!(
                    "el motivo {motivo:?} no tiene texto en {}: se vería la clave '{key}' cruda \
                     en pantalla",
                    locale.code()
                )
            });
            assert!(
                !texto.trim().is_empty(),
                "'{key}' está vacío en {}",
                locale.code()
            );
        }
    }
}

/// Cada motivo de fallo tiene **texto y siguiente paso** en los dos idiomas (`002`-FR-028).
///
/// El segundo no es adorno. FR-028 pide explicar *qué ocurrió* **y** *qué puede hacer el
/// usuario*: saber que el modelo tardó demasiado, sin que nadie diga que se puede probar con uno
/// más pequeño, deja a la persona igual de atascada que un `Failed` mudo.
///
/// El núcleo devuelve `SessionFailure::code()` y la piel elige la frase. Ese desacople permite
/// que el motivo siga el idioma activo (SC-009) y abre el hueco que este test cierra: un motivo
/// nuevo sin entrada en el catálogo se vería en pantalla como `session.failure.loquesea`.
#[test]
fn todo_motivo_de_fallo_tiene_texto_y_salida_en_ambos_idiomas() {
    use codify_core::domain::session::SessionFailure;

    for locale in [Locale::Es, Locale::En] {
        let entries = strings_for(locale).entries;
        for motivo in SessionFailure::all() {
            for sufijo in ["", ".next"] {
                let key = format!("session.failure.{}{sufijo}", motivo.code());
                let texto = entries.get(key.as_str()).unwrap_or_else(|| {
                    panic!(
                        "el motivo {motivo:?} no tiene '{key}' en {}: se vería la clave cruda en \
                         pantalla, que es justo lo que FR-028 prohíbe",
                        locale.code()
                    )
                });
                assert!(
                    !texto.trim().is_empty(),
                    "'{key}' está vacío en {}",
                    locale.code()
                );
            }
        }
    }
}

/// Cada estado de conexión tiene texto en **los dos** idiomas (`003`-T029).
///
/// Mismo patrón que `ProviderIssue` y `SessionFailure`: el núcleo devuelve `code()` y la piel
/// elige la frase. Ese desacople permite que el estado siga el idioma activo, y abre el hueco
/// que este test cierra — un estado nuevo sin entrada se vería como `connection.state.loquesea`.
#[test]
fn todo_estado_de_conexion_tiene_texto_en_ambos_idiomas() {
    use codify_core::application::connections::ConnectionState;

    for locale in [Locale::Es, Locale::En] {
        let entries = strings_for(locale).entries;
        for estado in ConnectionState::all() {
            let key = format!("connection.state.{}", estado.code());
            let texto = entries.get(key.as_str()).unwrap_or_else(|| {
                panic!(
                    "el estado {estado:?} no tiene texto en {}: se vería la clave cruda",
                    locale.code()
                )
            });
            assert!(
                !texto.trim().is_empty(),
                "'{key}' está vacío en {}",
                locale.code()
            );
        }
    }
}

/// `hidden` tiene que ocultar de verdad.
///
/// Este test nace de un defecto real: `#decide` llevaba `display: flex` y el atributo `hidden`
/// no hacía nada — el panel de decisión seguía visible cuando el código creía haberlo ocultado.
/// El atributo pone `display: none` desde la hoja del navegador, y **cualquier** `display` de
/// autor lo pisa. Basta una regla global para matar la clase entera.
#[test]
fn el_atributo_hidden_gana_siempre() {
    let css = read(ui_dir().join("styles.css")).replace(' ', "");
    let html = read(ui_dir().join("index.html"));

    assert!(
        html.contains("hidden"),
        "si nada usa `hidden`, este test sobra"
    );
    assert!(
        css.contains("[hidden]{display:none!important;}")
            || css.contains("[hidden]{display:none!important"),
        "falta la regla global `[hidden] {{ display: none !important; }}`: sin ella, cualquier \
         elemento con `display` propio ignora el atributo y el código que lo oculta no hace nada"
    );
}

/// Cada estado de sesión que el núcleo sabe reportar tiene texto en **los dos** idiomas.
///
/// Este test nace de un acoplamiento que estaba sin vigilar por los dos lados. La ventana pinta
/// el estado con `t(`session.state.${state}`)`, y ese `state` venía de `format!("{:?}")` sobre
/// la variante de Rust. Es decir: las claves del catálogo dependían **literalmente** del nombre
/// de la variante, y renombrar `Approved` habría dejado la interfaz mostrando la clave cruda.
///
/// El test de claves usadas no lo cubría —extrae `t("literal")`, y esto es una plantilla— y el
/// inverso pasaba por coincidencia de prefijo. Ahora el vínculo es explícito: el núcleo expone
/// un `code()` estable y esto comprueba que cada uno tenga su frase.
#[test]
fn todo_estado_de_sesion_tiene_texto_en_ambos_idiomas() {
    use codify_core::domain::session::SessionState;

    let estados = [
        SessionState::Ingesting,
        SessionState::Generating,
        SessionState::Refining,
        SessionState::Approved,
        SessionState::Failed,
        SessionState::Cancelled,
    ];

    for locale in [Locale::Es, Locale::En] {
        let entries = strings_for(locale).entries;
        for estado in estados {
            let key = format!("session.state.{}", estado.code());
            let texto = entries.get(key.as_str()).unwrap_or_else(|| {
                panic!(
                    "el estado {estado:?} no tiene texto en {}: la interfaz mostraría la clave \
                     cruda '{key}'",
                    locale.code()
                )
            });
            assert!(!texto.trim().is_empty(), "'{key}' está vacío");
        }
    }

    // Y los códigos son distintos entre sí: dos que colisionaran se verían como el mismo estado.
    let codigos: std::collections::HashSet<_> = estados.iter().map(|e| e.code()).collect();
    assert_eq!(
        codigos.len(),
        estados.len(),
        "hay códigos de estado repetidos"
    );
}

// ---------------------------------------------------------------------------
// S6 · Solo teclado
// ---------------------------------------------------------------------------

/// Un `tabindex` positivo saca al elemento de su sitio en el recorrido y rompe la promesa de
/// FR-025: que el orden de tabulación siga al orden visual.
#[test]
fn no_hay_tabindex_positivo() {
    for (name, content) in ui_files() {
        for marker in ["tabindex=\"", "tabIndex = ", "tabIndex="] {
            let mut rest = content.as_str();
            while let Some(at) = rest.find(marker) {
                rest = &rest[at + marker.len()..];
                let value: String = rest
                    .chars()
                    .skip_while(|c| c.is_whitespace())
                    .take_while(|c| c.is_ascii_digit() || *c == '-')
                    .collect();
                if let Ok(n) = value.parse::<i32>() {
                    assert!(
                        n <= 0,
                        "{name}: tabindex={n} rompe el orden natural de tabulación (FR-025)"
                    );
                }
            }
        }
    }
}

/// El foco tiene que dejar rastro. Sin esto, la navegación por teclado existe pero es
/// inutilizable: no se sabe dónde se está.
#[test]
fn el_foco_siempre_deja_rastro_visible() {
    let css = read(ui_dir().join("styles.css"));
    assert!(
        css.contains(":focus-visible"),
        "no hay estilo de foco: la navegación por teclado sería a ciegas (FR-025)"
    );

    let normalizado = css.replace(' ', "");
    for asesino in ["outline:none", "outline:0"] {
        assert!(
            !normalizado.contains(asesino),
            "`{asesino}` borra el indicador de foco (FR-025)"
        );
    }

    // El estilo de foco tiene que declarar un contorno real, no solo existir.
    let bloque = css
        .split(":focus-visible")
        .nth(1)
        .and_then(|r| r.split('}').next())
        .expect("la regla :focus-visible debe tener cuerpo");
    assert!(
        bloque.contains("outline:") && bloque.contains("px"),
        "la regla :focus-visible no declara un contorno con grosor: {bloque:?}"
    );
}

// ---------------------------------------------------------------------------
// S8 · Repositorio vacío
// ---------------------------------------------------------------------------

/// El repositorio vacío no puede presentarse como actividad rutinaria ni como final feliz:
/// necesita su propio tipo de bloque, su propio estado, y decir **qué hacer** (FR-019/FR-028).
#[test]
fn el_repositorio_vacio_tiene_presentacion_propia() {
    let main_js = read(ui_dir().join("main.js"));
    let stream_js = read(ui_dir().join("stream.js"));
    let css = read(ui_dir().join("styles.css"));

    assert!(
        main_js.contains("interviewMode"),
        "main.js debe reaccionar a `interviewMode`"
    );
    assert!(
        stream_js.contains("interview:"),
        "la corriente debe tener un tipo de bloque propio para el repositorio vacío, \
         no reutilizar `activity`"
    );
    assert!(
        css.contains("data-kind=\"interview\""),
        "el bloque de repositorio vacío debe distinguirse visualmente"
    );
    assert!(
        main_js.contains("session.interview_next"),
        "hay que decir qué hacer, no solo constatar que está vacío (FR-019/FR-028)"
    );

    let entries = strings_for(Locale::En).entries;
    assert!(
        entries.contains_key("session.state.interview"),
        "el estado no puede quedarse en «terminada» cuando no había nada que leer (FR-004)"
    );
}

/// Ninguna función de la interfaz se define y luego no se llama.
///
/// Nace de un hueco real: al cablear `provider.tier_degraded` (`001`-FR-018) comprobé que la
/// clave estuviera consumida y lo estaba —dentro de una función que nadie invocaba—. El test de
/// claves huérfanas no podía verlo: la cadena aparecía en el fichero. Una función muerta que
/// menciona una clave la mantiene «viva» a ojos de un contrato textual, y el usuario no ve nada.
#[test]
fn ninguna_funcion_de_la_interfaz_queda_sin_llamar() {
    let files = ui_files();
    let js: Vec<_> = files
        .iter()
        .filter(|(n, _)| n.ends_with(".js"))
        .cloned()
        .collect();
    // Los comentarios NO cuentan. Es un defecto real que este mismo test dejó pasar: al
    // cablear `renderFailure` (`002`-FR-028) quité su llamada para comprobar que el test la
    // atrapaba, y no lo hizo — porque un comentario mío de otra función mencionaba su nombre.
    // Una función muerta a la que la prosa mantiene viva sigue sin ejecutarse.
    let todo: String = js
        .iter()
        .map(|(_, c)| sin_comentarios(c))
        .collect::<Vec<_>>()
        .join("\n");

    let mut muertas = Vec::new();
    for (nombre, contenido) in &js {
        for linea in contenido.lines() {
            let t = linea.trim_start();
            let Some(resto) = t.strip_prefix("function ") else {
                continue;
            };
            let Some(fin) = resto.find('(') else { continue };
            let fn_name = &resto[..fin];
            if fn_name.is_empty() {
                continue;
            }
            // Se cuenta el identificador, no la llamada: `segments.map(renderSegment)` pasa la
            // función como valor y no lleva paréntesis. Contar `nombre(` la daba por muerta.
            // Con límite de palabra, para que `render` no se cuele dentro de `renderSegment`.
            let usos = ocurrencias_como_identificador(&todo, fn_name);
            if usos <= 1 {
                muertas.push(format!("{nombre}: {fn_name}"));
            }
        }
    }

    assert!(
        muertas.is_empty(),
        "funciones definidas y nunca llamadas — el código está, el usuario no lo ve: {muertas:?}"
    );
}

/// Cuenta apariciones de `nombre` como identificador completo, no como subcadena.
fn ocurrencias_como_identificador(fuente: &str, nombre: &str) -> usize {
    let bytes = fuente.as_bytes();
    let ident = |b: u8| b.is_ascii_alphanumeric() || b == b'_' || b == b'$';
    fuente
        .match_indices(nombre)
        .filter(|(i, _)| {
            let antes = *i == 0 || !ident(bytes[i - 1]);
            let fin = i + nombre.len();
            let despues = fin >= bytes.len() || !ident(bytes[fin]);
            antes && despues
        })
        .count()
}

/// Quita comentarios de JavaScript para que la prosa no cuente como uso de un identificador.
fn sin_comentarios(fuente: &str) -> String {
    let sin_bloque = strip_between(fuente, "/*", "*/");
    sin_bloque
        .lines()
        .map(|l| match l.find("//") {
            // `https://` no es un comentario; se descarta el falso positivo más común.
            Some(i) if i > 0 && l.as_bytes()[i - 1] == b':' => l,
            Some(i) => &l[..i],
            None => l,
        })
        .collect::<Vec<_>>()
        .join("\n")
}

// ---------------------------------------------------------------------------
// `004` · La superficie de conexión — forma propia, campos legibles, nombres únicos
// ---------------------------------------------------------------------------

/// Devuelve el cuerpo de la primera regla CSS cuyo selector empieza por `selector`.
fn regla_css(css: &str, selector: &str) -> Option<String> {
    let at = css.find(&format!("\n{selector} {{"))? + 1;
    let rest = &css[at..];
    let open = rest.find('{')? + 1;
    let close = rest.find('}')?;
    Some(rest[open..close].to_string())
}

/// Lee una declaración `prop: <n>ch` de un cuerpo de regla.
fn medida_ch(cuerpo: &str, prop: &str) -> Option<u32> {
    cuerpo
        .split(';')
        .map(str::trim)
        .find(|d| d.starts_with(prop) && d[prop.len()..].trim_start().starts_with(':'))
        .and_then(|d| d.split(':').nth(1))
        .and_then(|v| v.trim().strip_suffix("ch"))
        .and_then(|n| n.trim().parse().ok())
}

/// **SC-005** — los campos siguen siendo legibles con la ventana en su mínimo.
///
/// El criterio se expresa en `ch`, que es el ancho del carácter «0» de la fuente en uso: así
/// «24 caracteres» **es** la declaración, no una aproximación en píxeles que cambiaría de
/// significado al cambiar la tipografía. Los 147 px que había antes no decían nada sobre cuántos
/// caracteres cabían.
///
/// El suelo por sí solo no basta: dentro de un contenedor que no envuelve produce desbordamiento
/// en vez de campos legibles, y sin techo un valor largo empuja a los demás fuera de la fila.
#[test]
fn los_campos_del_formulario_caben_en_la_ventana_minima() {
    const MINIMO: u32 = 24;
    // Sin comentarios: las declaraciones van comentadas y un `/* … */` delante de `min-width`
    // haría que el análisis no la viera.
    let css = strip_between(&read(ui_dir().join("styles.css")), "/*", "*/");

    let campo = regla_css(&css, ".campo")
        .expect("styles.css debe declarar una regla `.campo` para los campos del formulario");

    let min = medida_ch(&campo, "min-width").unwrap_or_else(|| {
        panic!("`.campo` debe declarar `min-width` en `ch`: SC-005 se mide en caracteres, no en píxeles")
    });
    assert!(
        min >= MINIMO,
        "`.campo` declara `min-width: {min}ch` y SC-005 exige al menos {MINIMO}: por debajo, una \
         dirección de proveedor deja de ser adivinable"
    );

    assert!(
        medida_ch(&campo, "max-width").is_some(),
        "`.campo` debe declarar también un `max-width` en `ch`: el suelo no impide que un valor \
         largo empuje a los demás campos fuera de la fila"
    );

    let contenedor = regla_css(&css, ".alta-campos")
        .expect("styles.css debe declarar `.alta-campos`, el contenedor de los campos");
    assert!(
        contenedor.replace(' ', "").contains("flex-wrap:wrap"),
        "`.alta-campos` debe envolver: un `min-width` dentro de un contenedor que no envuelve no \
         da campos legibles, da desbordamiento horizontal — la otra mitad de SC-005"
    );
}

/// **FR-001, FR-002a, FR-004** — la superficie tiene forma propia.
///
/// Nace de mirar la aplicación: `#conexiones` reutilizaba la clase de la barra de estado del
/// proveedor, que es una fila pensada para información pasiva, y metía dentro un formulario de
/// tres campos. Todos los tests pasaban y la pantalla no se entendía.
///
/// FR-006 exige, para los nombres de región, que la comprobación no «dependa de que alguien lo
/// note mirando». Dejar la forma al ojo sería aplicar ese criterio a la mitad del spec — y el ojo
/// ya falló aquí una vez.
#[test]
fn la_superficie_de_conexion_tiene_forma_propia() {
    let html = read(ui_dir().join("index.html"));

    let etiqueta = html
        .split_once("<section id=\"conexiones\"")
        .and_then(|(_, rest)| rest.split_once('>').map(|(tag, _)| tag.to_string()))
        .expect("index.html debe tener la sección `#conexiones`");
    assert!(
        !etiqueta.contains("provider"),
        "`#conexiones` sigue llevando la presentación de la barra de estado del proveedor \
         (`{}`). FR-001 prohíbe reutilizar la disposición de una barra de estado para contener \
         un formulario",
        etiqueta.trim()
    );

    let seccion = html
        .split_once("<section id=\"conexiones\"")
        .map(|(_, rest)| {
            rest.split_once("</section>")
                .map(|(s, _)| s)
                .unwrap_or(rest)
        })
        .expect("la sección `#conexiones` debe cerrar");

    let abre_plegable = seccion
        .find("<details")
        .expect("FR-002a: el formulario de conexión debe vivir en un contenedor plegable");
    let tag_details = seccion[abre_plegable..]
        .split_once('>')
        .map(|(t, _)| t)
        .unwrap_or_default();
    assert!(
        !tag_details.contains(" open"),
        "el plegable del formulario arranca abierto, así que el caso por defecto —sin cuentas \
         conectadas— sigue pagando por el raro (FR-002a)"
    );

    let lista = seccion
        .find("id=\"conexiones-lista\"")
        .expect("la lista de cuentas conectadas debe estar en la sección");
    assert!(
        lista < abre_plegable,
        "la lista de cuentas está **dentro** del plegable del formulario: al plegarlo desaparecen \
         las cuentas conectadas, y FR-004 pide justo distinguir una cosa de la otra"
    );
}

/// **FR-005, FR-006, SC-002** — dos regiones no pueden llamarse igual.
///
/// Se comprueban las claves **y los textos resueltos**, en los dos idiomas. Solo las claves
/// dejaría pasar dos claves distintas con el mismo texto; solo los textos dejaría pasar el caso
/// en que un idioma los distingue y el otro no. Quien usa un lector de pantalla oye el texto.
///
/// Al escribirlo había dos pares en conflicto, y llevaban ahí desde `003` sin que nadie los
/// notara mirando: `#conexiones` sonaba igual que `#provider`, y el pie igual que la barra.
#[test]
fn ningun_par_de_regiones_comparte_nombre_accesible() {
    let html = read(ui_dir().join("index.html"));

    // Un `<dialog>` no es una región. Se nombra solo, con su `aria-labelledby`, y forzarle un rol
    // de landmark lo cuela en el rotor **estando cerrado**: quien navega por regiones llega a una
    // entrada que apunta a un espacio invisible. Lo dijo una persona probándolo, y este test
    // pasaba mientras tanto porque contaba esa región como legítima.
    let dialogos_con_rol: Vec<String> = html
        .split("<dialog")
        .skip(1)
        .filter_map(|rest| rest.split_once('>').map(|(tag, _)| tag.to_string()))
        .filter(|tag| tag.contains("role=\"region\"") || tag.contains("data-i18n-aria"))
        .collect();
    assert!(
        dialogos_con_rol.is_empty(),
        "un `<dialog>` con rol de región aparece en el rotor aunque esté cerrado, y pierde su \
         propia semántica de diálogo:\n{}",
        dialogos_con_rol.join("\n")
    );

    // El censo cuenta las dos formas de nombrar una región: la clave directa y la referencia a un
    // encabezado visible. La segunda es preferible —el nombre tiene un solo dueño— y dejarla fuera
    // haría que este test dejara de ver justo las regiones mejor construidas.
    let mut claves: Vec<String> = html
        .split("data-i18n-aria=\"")
        .skip(1)
        .filter_map(|rest| rest.split('"').next().map(str::to_string))
        .collect();
    for rest in html.split("aria-labelledby=\"").skip(1) {
        let Some(id) = rest.split('"').next() else {
            continue;
        };
        let Some((_, tras_id)) = html.split_once(&format!("id=\"{id}\"")) else {
            continue;
        };
        let Some(tag) = tras_id.split_once('>').map(|(t, _)| t) else {
            continue;
        };
        if let Some(clave) = between(tag, "data-i18n=\"", "\"") {
            claves.push(clave);
        }
    }
    assert!(
        claves.len() >= 2,
        "no se encontraron regiones con nombre accesible en index.html"
    );

    let mut repetidas = Vec::new();
    for (i, a) in claves.iter().enumerate() {
        if claves[..i].contains(a) && !repetidas.contains(a) {
            repetidas.push(a.clone());
        }
    }
    assert!(
        repetidas.is_empty(),
        "estas claves nombran más de una región, así que suenan igual al recorrerlas: {repetidas:?}"
    );

    for locale in [Locale::Es, Locale::En] {
        let entries = strings_for(locale).entries;
        let mut vistos: Vec<(String, String)> = Vec::new();
        for clave in &claves {
            let texto = entries.get(clave.as_str()).unwrap_or_else(|| {
                panic!(
                    "la región '{clave}' no tiene texto en {}: se oiría la clave cruda",
                    locale.code()
                )
            });
            if let Some((otra, _)) = vistos.iter().find(|(_, t)| t == texto) {
                panic!(
                    "en {} las regiones '{otra}' y '{clave}' se llaman igual ({texto:?}): quien \
                     navega por regiones no puede saber en cuál está",
                    locale.code()
                );
            }
            vistos.push((clave.clone(), texto.to_string()));
        }
    }
}

/// Nombre de la función que envuelve la posición `at` dentro de `fuente`.
///
/// Heurística deliberada: la `function nombre(` más cercana hacia atrás. Es suficiente para un
/// JavaScript sin anidamiento profundo, y decir que dos escrituras «están en la misma función»
/// necesita exactamente esto.
fn funcion_que_envuelve(fuente: &str, at: usize) -> String {
    fuente[..at]
        .rmatch_indices("function ")
        .next()
        .map(|(i, _)| {
            fuente[i + "function ".len()..]
                .chars()
                .take_while(|c| c.is_alphanumeric() || *c == '_' || *c == '$')
                .collect()
        })
        .unwrap_or_else(|| "<nivel de módulo>".to_string())
}

/// Todas las posiciones de `aguja` en cada archivo `.js` de la interfaz.
fn sitios(aguja: &str) -> Vec<(String, String)> {
    let mut out = Vec::new();
    for (name, content) in ui_files() {
        if !name.ends_with(".js") {
            continue;
        }
        let sin = sin_comentarios(&content);
        let mut desde = 0;
        while let Some(rel) = sin[desde..].find(aguja) {
            let at = desde + rel;
            out.push((name.clone(), funcion_que_envuelve(&sin, at)));
            desde = at + aguja.len();
        }
    }
    out
}

/// **FR-003a, SC-006** — las dos superficies del modo no pueden discrepar.
///
/// El modo se enseña en dos sitios porque responden preguntas distintas: la insignia dice en qué
/// modo estás de un vistazo, la casilla permite cambiarlo. Lo que no puede ocurrir es que digan
/// cosas distintas.
///
/// Al escribir este test había **dos escritores y ninguna sincronización**: la insignia se pintaba
/// desde `ui.local`, que valía `true` y no se reasignaba jamás, y la casilla vivía por su cuenta
/// en el DOM. La insignia decía «local» hicieras lo que hicieras.
///
/// La regla, entonces: un solo escritor por superficie, y el mismo para las dos. Contarlos es lo
/// que lo impide; compararlos después solo lo detectaría.
#[test]
fn el_modo_no_puede_discrepar_entre_sus_dos_superficies() {
    let insignia = sitios("dataset.mode =");
    let casilla = sitios(".checked =");

    assert_eq!(
        insignia.len(),
        1,
        "la insignia de modo se escribe en {} sitios: {insignia:?}. Con más de uno pueden \
         discrepar; con ninguno, no se pinta",
        insignia.len()
    );
    assert_eq!(
        casilla.len(),
        1,
        "la casilla de modo se escribe en {} sitios: {casilla:?}. Si nunca se escribe, refleja lo \
         que el usuario pulsó y no lo que el backend guardó",
        casilla.len()
    );
    assert_eq!(
        insignia[0], casilla[0],
        "las dos superficies del modo se pintan en sitios distintos ({:?} y {:?}): existe un \
         camino por el que una cambia y la otra no",
        insignia[0], casilla[0]
    );
}

/// El módulo que es dueño del modo. Nadie más puede tener una idea propia de cuál es.
const DUENO_DEL_MODO: &str = "mode.js";

/// **FR-003a** — la interfaz no guarda copia del modo.
///
/// Sin esto, el test anterior pasa con las dos superficies pintadas a la vez desde un valor
/// equivocado — que es literalmente lo que ocurría: `ui.local` valía `true` para siempre, y
/// `start_session` decidía el modo de la sesión leyendo esa copia en vez del estado guardado. El
/// resultado es que `set_mode` escribía donde nadie leía y `003`-FR-008a no se cumplía.
///
/// La fuente única es el backend. La interfaz pregunta; no recuerda.
#[test]
fn la_interfaz_no_tiene_su_propia_idea_del_modo() {
    let mut copias = Vec::new();
    for (name, content) in ui_files() {
        if !name.ends_with(".js") || name == DUENO_DEL_MODO {
            continue;
        }
        for (n, linea) in sin_comentarios(&content).lines().enumerate() {
            if linea.contains("local:") {
                copias.push(format!("{name}:{}: {}", n + 1, linea.trim()));
            }
        }
    }
    assert!(
        copias.is_empty(),
        "estos módulos llevan su propia idea del modo, y una copia que nadie sincroniza es una \
         copia que miente:\n{}",
        copias.join("\n")
    );

    let dueno = ui_files()
        .into_iter()
        .find(|(n, _)| n == DUENO_DEL_MODO)
        .map(|(_, c)| sin_comentarios(&c))
        .unwrap_or_else(|| panic!("falta `ui/{DUENO_DEL_MODO}`, que es quien posee el modo"));
    let mutable: Vec<&str> = dueno
        .lines()
        .filter(|l| l.starts_with("let ") || l.starts_with("var "))
        .collect();
    assert!(
        mutable.is_empty(),
        "`{DUENO_DEL_MODO}` guarda estado mutable de módulo ({mutable:?}): sería una caché del \
         modo, y una caché es una segunda verdad esperando a divergir"
    );
}

/// Comandos del backend que **hoy** no tiene quien los llame, y por qué se toleran.
///
/// Cada uno con su issue. La lista no es un permiso: es una deuda con nombre, y el test existe
/// para que no crezca sola.
const COMANDOS_SIN_LLAMAR: &[(&str, &str)] = &[
    // El flujo de conexión se abre y nunca se cierra: `connect_provider` devuelve un desafío, la
    // interfaz enseña una frase y no hay campo donde escribir la credencial. Encontrado probando
    // con una persona delante, que pulsó «Conectar» y no ocurrió nada — correctamente.
    ("complete_connection", "#54"),
    // Las propuestas llegan por evento. Este comando las devolvería al recargar, y nadie lo
    // invoca: o sobra, o falta el camino de recuperación. Encontrado por este mismo test.
    ("pending_proposals", "sin issue todavía"),
];

/// Un comando registrado que nadie invoca compila, pasa sus tests y no hace nada.
///
/// Es el hueco de #48 en su forma más pura, y volvió a morder: `complete_connection` llevaba
/// desde `003` sin quien lo llamara, así que conectar un proveedor era **imposible** mientras
/// todos los tests estaban verdes. `ninguna_funcion_de_la_interfaz_queda_sin_llamar` no lo veía
/// porque mira funciones de JavaScript, no comandos.
#[test]
fn ningun_comando_del_backend_queda_sin_invocar() {
    let rust = read(crate_dir().join("src/commands.rs"));
    let comandos: Vec<String> = rust
        .split("#[tauri::command]")
        .skip(1)
        .filter_map(|rest| rest.split_once(" fn "))
        .filter_map(|(_, tras)| {
            let nombre: String = tras
                .chars()
                .take_while(|c| c.is_alphanumeric() || *c == '_')
                .collect();
            (!nombre.is_empty()).then_some(nombre)
        })
        .collect();
    assert!(
        comandos.len() > 10,
        "no se reconocieron los comandos de commands.rs: el análisis está roto, no el código"
    );

    let js: String = ui_files()
        .into_iter()
        .filter(|(n, _)| n.ends_with(".js"))
        .map(|(_, c)| sin_comentarios(&c))
        .collect();

    let mut huerfanos = Vec::new();
    for cmd in &comandos {
        if js.contains(&format!("invoke(\"{cmd}\"")) {
            continue;
        }
        match COMANDOS_SIN_LLAMAR.iter().find(|(n, _)| n == cmd) {
            Some(_) => {}
            None => huerfanos.push(cmd.clone()),
        }
    }
    assert!(
        huerfanos.is_empty(),
        "estos comandos están registrados y ningún JavaScript los invoca, así que la parte de la \
         aplicación que sirven es inalcanzable: {huerfanos:?}"
    );

    let saldados: Vec<&str> = COMANDOS_SIN_LLAMAR
        .iter()
        .filter(|(n, _)| js.contains(&format!("invoke(\"{n}\"")))
        .map(|(n, _)| *n)
        .collect();
    assert!(
        saldados.is_empty(),
        "estos comandos ya se invocan y siguen en la lista de deuda: quitarlos de ahí es parte de \
         darlos por cerrados ({saldados:?})"
    );
}
