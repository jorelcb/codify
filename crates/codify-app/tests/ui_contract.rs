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
        let usada = files.iter().any(|(_, c)| c.contains(key))
            || rust.contains(key)
            || key.rsplit_once('.').is_some_and(|(prefix, _)| {
                let plantilla = format!("{prefix}.${{");
                files.iter().any(|(_, c)| c.contains(&plantilla))
            });
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
    let todo: String = js
        .iter()
        .map(|(_, c)| c.as_str())
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
